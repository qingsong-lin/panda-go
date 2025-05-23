package panda_nacos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	nacos_log "github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/common/nacos_server"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"panda-go/common/log"
	"panda-go/common/utils"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const (
	NACOS_ENV_CONFIG_KEY = "NACOS_CONFIG"
	SERVICE_NAME         = "JAEGER_SERVICE_NAME"
	NAMESPACE            = "NAMESPACE"
	HOST_NAME            = "HOSTNAME"
)

type Handler func(key, data string)

type NacosClient struct {
	configClient         config_client.IConfigClient
	ConfigCacheMap       sync.Map
	ConfigWatchHandleMap sync.Map
	nacosServer          *nacos_server.NacosServer
}

var (
	innerNacosClient  *NacosClient
	prometheusMonitor = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "nacos_get_config",
		Help: "nacos_get_config",
	}, []string{"key", "group"})
	NamespaceEnv = os.Getenv(NAMESPACE)
	one          sync.Once
	ServerName   = os.Getenv(SERVICE_NAME)
	PodName      = os.Getenv(HOST_NAME)
)

func init() {
	//var nacosConfig string
	//for {
	//	nacosConfig = os.Getenv(NACOS_ENV_CONFIG_KEY)
	//	if nacosConfig == "" {
	//		log.Info(context.TODO(), "[Init] nacos config  is emtpy")
	//		time.Sleep(200 * time.Millisecond)
	//		continue
	//	}
	//	break
	//}
	//nacosParam := vo.NacosClientParam{}
	//err := json.Unmarshal([]byte(nacosConfig), &nacosParam)
	//if err != nil {
	//	log.Fatal(context.TODO(), "[Init] json.Unmarshal Nacos Config failed", zap.String("NACOS_CONFIG", nacosConfig), zap.Error(err))
	//}
	//for {
	//	if utils.Ping(nacosParam.ServerConfigs[0].IpAddr) {
	//		log.Info(context.TODO(), fmt.Sprintf("Ping to %s successful!\n", nacosParam.ServerConfigs[0].IpAddr))
	//		break
	//	} else {
	//		log.Info(context.TODO(), fmt.Sprintf("Waiting for %s to respond...\n", nacosParam.ServerConfigs[0].IpAddr))
	//		time.Sleep(200 * time.Millisecond)
	//	}
	//}
	// 等待istio相应资源准备就绪
	time.Sleep(3 * time.Second)

	if ServerName == "" {
		ServerName = "common"
	}
	if NamespaceEnv == "" {
		NamespaceEnv = "DEFAULT_GROUP"
	}
}

type ConfigParam struct {
	DataId           string `param:"dataId"`  //required
	Group            string `param:"group"`   //required
	Content          string `param:"content"` //required
	Tag              string `param:"tag"`
	AppName          string `param:"appName"`
	BetaIps          string `param:"betaIps"`
	CasMd5           string `param:"casMd5"`
	Type             string `param:"type"`
	SrcUser          string `param:"srcUser"`
	EncryptedDataKey string `param:"encryptedDataKey"`
	KmsKeyId         string `param:"kmsKeyId"`
	UsageType        string `param:"usageType"`
}

func GetNacosClient() *NacosClient {
	one.Do(func() {
		//prometheus.Register(prometheusMonitor)
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		nacos_log.SetLogger(log.GetLogger().With(zap.String("app", "[NACOS]")).Sugar())
		nacosConfig := os.Getenv(NACOS_ENV_CONFIG_KEY)
		log.Info(ctx, "[Init] nacos config", zap.String("nacosConfig", nacosConfig))
		nacosParam := vo.NacosClientParam{}
		err := json.Unmarshal([]byte(nacosConfig), &nacosParam)
		if err != nil {
			log.Fatal(ctx, "[GetNacosClient] json.Unmarshal Nacos Config failed", zap.String("NACOS_CONFIG", nacosConfig), zap.Error(err))
		}
		//log.Info(ctx, "[Init] nacos config", zap.Any("nacosConfig", nacosParam))
		if nacosParam.ClientConfig != nil {
			nacosParam.ClientConfig.NotLoadCacheAtStart = true
			if //goland:noinspection ALL
			nacosParam.ClientConfig.CacheDir == "" && runtime.GOOS != "windows" {
				nacosParam.ClientConfig.CacheDir = fmt.Sprintf("/tmp/nacos/cache/%s-%d", nacosParam.ClientConfig.NamespaceId, time.Now().UnixNano())
			}

			if //goland:noinspection ALL
			nacosParam.ClientConfig.LogDir == "" && runtime.GOOS != "windows" {
				nacosParam.ClientConfig.LogDir = fmt.Sprintf("/tmp/nacos/log/%s-%d", nacosParam.ClientConfig.NamespaceId, time.Now().UnixNano())
			}
			nacosParam.ClientConfig.AppName = ServerName
			nacosParam.ClientConfig.UpdateCacheWhenEmpty = true

		} else {
			log.Fatal(ctx, "[GetNacosClient] nacosParam.ClientConfig is nil")
		}

		if len(nacosParam.ServerConfigs) == 0 {
			log.Fatal(ctx, "[GetNacosClient] nacosParam.ServerConfigs is empty")
		}
		if nacosParam.ClientConfig.BeatInterval == 0 {
			nacosParam.ClientConfig.BeatInterval = 5000
		}
		if nacosParam.ClientConfig.TimeoutMs == 0 {
			nacosParam.ClientConfig.TimeoutMs = 10000
		}

		// 创建 Nacos 客户端
		client, err := clients.NewConfigClient(nacosParam)
		if err != nil {
			log.Fatal(ctx, "[GetNacosClient] NewConfigClient failed", zap.Error(err))
		}

		nacosConfigClientex := client.(*config_client.ConfigClient)
		httpAgent, _ := nacosConfigClientex.GetHttpAgent()
		clientConf, _ := nacosConfigClientex.GetClientConfig()
		//namingHeader := map[string][]string{
		//	"Client-Version": {constant.CLIENT_VERSION},
		//	"User-Agent":     {constant.CLIENT_VERSION},
		//	"RequestId":      {uid.String()},
		//	"Request-Module": {"Naming"},
		//}
		nacosServer, err := nacos_server.NewNacosServer(ctx, nacosParam.ServerConfigs, clientConf, httpAgent, clientConf.TimeoutMs, clientConf.Endpoint, nil)
		if err != nil {
			log.Fatal(ctx, "[GetNacosClient] NewNacosServer failed", zap.Error(err))
		}
		innerNacosClient = &NacosClient{
			configClient: client,
			nacosServer:  nacosServer,
		}
		// 定时五分钟刷新：有人反馈nacos出现订阅不更新的情况，五分钟强制刷新
		go func() {
			innerNacosClient.RefreshConfig(ctx)
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		// 创建一个 goroutine 来等待信号
		go func() {
			defer cancel()
			<-c
			client.CloseClient()
			log.Info(ctx, "accept quit signal, close global nacos client!")
			// 执行清理操作的代码
		}()
	})
	if innerNacosClient == nil {
		log.Fatal(context.TODO(), "GetNacosClient fail, innerNacosClient is nil")
	}
	return innerNacosClient
}

func changeOwnConfigParam2vo(inParam ConfigParam) vo.ConfigParam {
	return vo.ConfigParam{
		DataId:           inParam.DataId,
		Group:            inParam.Group,
		Content:          inParam.Content,
		Tag:              inParam.Tag,
		AppName:          inParam.AppName,
		BetaIps:          inParam.BetaIps,
		CasMd5:           inParam.CasMd5,
		Type:             inParam.Type,
		SrcUser:          inParam.SrcUser,
		EncryptedDataKey: inParam.EncryptedDataKey,
		KmsKeyId:         inParam.KmsKeyId,
		UsageType:        vo.UsageType(inParam.UsageType),
	}
}

// PublishConfig oldData 不为空时候确保原子发布配置, 必须保证上一次的状态没有被变更过才去修改对应配置
func (c *NacosClient) PublishConfig(ctx context.Context, configParam *ConfigParam, oldData string) (success bool, err error) {
	err = retry.Do(func() error {
		if oldData != "" {
			configParam.CasMd5 = utils.StringToMd5String(oldData)
		}
		if configParam.Group == "" {
			configParam.Group = NamespaceEnv
		}
		if configParam.AppName == "" {
			configParam.AppName = ServerName
		}
		success, err = c.configClient.PublishConfig(changeOwnConfigParam2vo(*configParam))
		if err != nil {
			log.Error(ctx, "[NacosClient] PublishConfig err", zap.Error(err))
			return err
		}
		if success {
			log.Info(ctx, "[NacosClient] PublishConfig success", zap.Any("param", configParam))
		}
		return nil
	}, retry.Delay(2*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Attempts(2),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			log.Error(ctx, "[NacosClient] PublishConfig error", zap.String("key", configParam.DataId), zap.String("value", configParam.Content), zap.String("group", configParam.Group), zap.Error(err))
		}))
	return
}

func (c *NacosClient) GetAndWatchConfig(ctx context.Context, key string, handler Handler) error {
	param := &ConfigParam{
		Group:  NamespaceEnv,
		DataId: key,
	}
	cfg, err := c.GetConfig(ctx, param)
	if nil != err {
		log.Error(ctx, "GetConfig fail.", zap.String("key", key), zap.String("Group", NamespaceEnv), zap.Error(err))
	} else {
		handler(key, cfg)
		log.Debug(ctx, "GetConfig success.", zap.String("key", key), zap.String("cfg", cfg))
	}
	c.WatchConfig(ctx, param, handler)
	return err
}

func (c *NacosClient) WatchConfig(ctx context.Context, inParam *ConfigParam, handler Handler) (err error) {
	configParamEntity := changeOwnConfigParam2vo(*inParam)
	configParam := &configParamEntity
	err = retry.Do(func() error {
		if configParam == nil {
			return errors.New("[NacosClient.WatchConfig]configParam is nil")
		}
		if configParam.Group == "" {
			configParam.Group = NamespaceEnv
		}
		configParam.OnChange = func(namespace, group, dataId, data string) {
			log.Debug(ctx, "[NacosClient] config has changed", zap.String("key", configParam.DataId), zap.String("value", configParam.Content), zap.String("group", configParam.Group), zap.Any("data", data))
			handler(dataId, data)
			c.ConfigCacheMap.Store(*inParam, data)
		}
		if err := c.configClient.ListenConfig(*configParam); err != nil {
			log.Error(ctx, "[NacosClient] In WatchConfig, configClient.ListenConfig fail", zap.Error(err))
			return err
		}
		c.ConfigWatchHandleMap.Store(*inParam, handler)
		return nil
	}, retry.Delay(2*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Attempts(2),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			log.Error(ctx, "[NacosClient] WatchConfig error", zap.Any("inParam", inParam), zap.Error(err))
		}))
	return
}

func (c *NacosClient) GetConfig(ctx context.Context, inParam *ConfigParam) (data string, err error) {
	configParamEntity := changeOwnConfigParam2vo(*inParam)
	configParam := &configParamEntity
	err = retry.Do(func() error {
		if configParam == nil {
			return errors.New("[NacosClient.GetConfig]configParam is nil")
		}
		if configParam.Group == "" {
			configParam.Group = NamespaceEnv
		}
		content, err := c.configClient.GetConfig(*configParam)
		if err != nil {
			return err
		}
		data = content
		log.Debug(ctx, "Nacos: GetConfig success", zap.Any("inParam", inParam))
		c.ConfigCacheMap.Store(*inParam, content)
		return nil
	},
		retry.Delay(2*time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.Attempts(2),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			log.Error(ctx, "Nacos: GetConfig error", zap.Any("inParam", inParam), zap.Error(err))
		}),
	)
	// todo
	//prometheusMonitor.WithLabelValues(configParam.DataId, configParam.Group).Inc()
	return
}

func (c *NacosClient) RefreshConfig(ctx context.Context) {
	go func() {
		ticker := time.Tick(10 * time.Second)
		for {
			select {
			case <-ctx.Done():
				log.Info(ctx, "RefreshConfig ctx done", zap.Error(ctx.Err()))
				return
			case <-ticker:
				c.ConfigWatchHandleMap.Range(func(key, value any) bool {
					configParam := key.(ConfigParam)
					handler := value.(Handler)
					log.Debug(ctx, "Nacos: RefreshConfig start", zap.Any("configParam", configParam))
					data, err := c.GetConfig(ctx, &configParam)
					if err != nil {
						return true
					}
					oldData, ok := c.ConfigCacheMap.Load(configParam)
					if ok {
						if reflect.DeepEqual(oldData, data) {
							return true
						}
					}
					log.Debug(ctx, "Nacos: RefreshConfig success", zap.Any("configParam", configParam))
					c.ConfigCacheMap.Store(configParam, data)
					handler(configParam.DataId, data)
					return true
				})
			}
		}
	}()
}

type Namespace struct {
	Namespace         string `json:"namespace"`
	NamespaceShowName string `json:"namespaceShowName"`
	Quota             int    `json:"quota"`
	ConfigCount       int    `json:"configCount"`
	Type              int8   `json:"type"`
}

type SearchNamespaceRsp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    []Namespace `json:"data"`
}

type SearchListenConfigRsp struct {
	CollectStatus           int               `json:"collectStatus"`
	LisentersGroupkeyStatus map[string]string `json:"lisentersGroupkeyStatus"`
}

type HistoryConfigItem struct {
	Id               string `json:"id"`
	LastId           int32  `json:"lastId"`
	DataId           string `json:"dataId"`
	Group            string `json:"group"`
	Tenant           string `json:"tenant"`
	AppName          string `json:"appName"`
	Md5              string `json:"md5"`
	Content          string `json:"content"`
	SrcIp            string `json:"srcIp"`
	SrcUser          string `json:"srcUser"`
	OpType           string `json:"opType"`
	CreatedTime      string `json:"createdTime"`
	LastModifiedTime string `json:"lastModifiedTime"`
}

type SearchHistoryNacosConfigRsp struct {
	TotalCount     int32               `json:"totalCount"`
	PageNumber     int32               `json:"pageNumber"`
	PagesAvailable int32               `json:"pagesAvailable"`
	PageItems      []HistoryConfigItem `json:"pageItems"`
}

// GetHistoryNacosConfigDetail 根据nid查询历史数据详情
func (c *NacosClient) GetHistoryNacosConfigDetail(ctx context.Context, params map[string]string) (*HistoryConfigItem, error) {
	if _, ok := params["nid"]; !ok {
		return nil, errors.New("请输入nid")
	}
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.CONFIG_BASE_PATH+"/history", params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: SearchHistoryNacosConfig error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: SearchHistoryNacosConfig success", zap.Any("param", params))
	var historyNacosConfig HistoryConfigItem
	err = json.Unmarshal([]byte(content), &historyNacosConfig)
	if err != nil {
		return nil, err
	}
	return &historyNacosConfig, nil
}

// SearchHistoryNacosConfig 查询历史数据条目
func (c *NacosClient) SearchHistoryNacosConfig(ctx context.Context, params map[string]string) (*SearchHistoryNacosConfigRsp, error) {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	params["search"] = "accurate"
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.CONFIG_BASE_PATH+"/history", params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: SearchHistoryNacosConfig error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: SearchHistoryNacosConfig success", zap.Any("param", params))
	var historyNacosConfigs SearchHistoryNacosConfigRsp
	err = json.Unmarshal([]byte(content), &historyNacosConfigs)
	if err != nil {
		return nil, err
	}
	return &historyNacosConfigs, nil
}

// SearchListenConfig 监听查询
func (c *NacosClient) SearchListenConfig(ctx context.Context, params map[string]string, isIp bool) (*SearchListenConfigRsp, error) {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	path := constant.CONFIG_LISTEN_PATH
	if isIp {
		path = constant.CONFIG_BASE_PATH + "/listener"
	}
	content, err := c.nacosServer.ReqConfigApi(path, params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: SearchListenConfig error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: SearchListenConfig success", zap.Any("param", params))
	var listenConfig SearchListenConfigRsp
	err = json.Unmarshal([]byte(content), &listenConfig)
	if err != nil {
		return nil, err
	}
	return &listenConfig, nil
}

// SearchNamespaces 查询nacos下的所有命名空间
func (c *NacosClient) SearchNamespaces(ctx context.Context) (*SearchNamespaceRsp, error) {
	params := map[string]string{}
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey

	content, err := c.nacosServer.ReqConfigApi(constant.NAMESPACE_PATH, params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: SearchNamespaces error", zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: SearchNamespaces success")
	var namespacePage SearchNamespaceRsp
	err = json.Unmarshal([]byte(content), &namespacePage)
	if err != nil {
		return nil, err
	}
	return &namespacePage, nil
}

type NacosService struct {
	Name                 string `json:"name"`
	GroupName            string `json:"groupName"`
	ClusterCount         int32  `json:"clusterCount"`
	IpCount              int32  `json:"ipCount"`
	HealthyInstanceCount int32  `json:"healthyInstanceCount"`
	TriggerFlag          string `json:"triggerFlag"`
}

type SearchNacosServiceListRsp struct {
	Count       int32          `json:"count"`
	ServiceList []NacosService `json:"serviceList"`
}

func (c *NacosClient) SearchNacosServiceList(ctx context.Context, params map[string]string) (*SearchNacosServiceListRsp, error) {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_BASE_PATH+"/catalog/services", params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: SearchNacosServiceList error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: SearchHistoryNacosConfig success", zap.Any("param", params))
	var searchNacosServiceList SearchNacosServiceListRsp
	err = json.Unmarshal([]byte(content), &searchNacosServiceList)
	if err != nil {
		return nil, err
	}
	return &searchNacosServiceList, nil
}

type GetNacosServiceInfoRsp struct {
	Service struct {
		Name             string  `json:"name"`
		GroupName        string  `json:"groupName"`
		ProtectThreshold float64 `json:"protectThreshold"`
		Selector         struct {
			Type string `json:"type"`
		} `json:"selector"`
		Metadata map[string]interface{} `json:"metadata"`
	} `json:"service"`
	Clusters []struct {
		ServiceName   string `json:"serviceName"`
		Name          string `json:"name"`
		HealthChecker struct {
			Type string `json:"type"`
		} `json:"healthChecker"`
		DefaultPort      int32                  `json:"defaultPort"`
		DefaultCheckPort int32                  `json:"defaultCheckPort"`
		UseIPPort4Check  bool                   `json:"useIPPort4Check"`
		Metadata         map[string]interface{} `json:"metadata"`
	} `json:"clusters"`
}

func (c *NacosClient) GetNacosServiceInfo(ctx context.Context, params map[string]string) (*GetNacosServiceInfoRsp, error) {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_BASE_PATH+"/catalog/service", params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: GetNacosServiceInfo error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: GetNacosServiceInfo success", zap.Any("param", params))
	var getNacosServiceInfoRsp GetNacosServiceInfoRsp
	err = json.Unmarshal([]byte(content), &getNacosServiceInfoRsp)
	if err != nil {
		return nil, err
	}
	return &getNacosServiceInfoRsp, nil
}

type NacosServiceInstance struct {
	Ip                        string                 `json:"ip"`
	Port                      int32                  `json:"port"`
	Weight                    float64                `json:"weight"`
	Healthy                   bool                   `json:"healthy"`
	Enabled                   bool                   `json:"enabled"`
	Ephemeral                 bool                   `json:"ephemeral"`
	ClusterName               string                 `json:"clusterName"`
	ServiceName               string                 `json:"serviceName"`
	Metadata                  map[string]interface{} `json:"metadata"`
	InstanceHeartBeatInterval int32                  `json:"instanceHeartBeatInterval"`
	InstanceHeartBeatTimeOut  int32                  `json:"instanceHeartBeatTimeOut"`
	IpDeleteTimeout           int32                  `json:"ipDeleteTimeout"`
}

func (c *NacosClient) UpdateNacosServiceInstance(ctx context.Context, params map[string]string) error {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_PATH, params, headers, http.MethodPut, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: UpdateNacosServiceInstance error", zap.Any("param", params), zap.Error(err))
		return err
	}
	log.Info(ctx, "Nacos: UpdateNacosServiceInstance success", zap.Any("param", params))
	if content != "ok" {
		return errors.New(content)
	}
	return nil
}

type ListNacosServiceInstanceRsp struct {
	List  []NacosServiceInstance `json:"list"`
	Count int32                  `json:"count"`
}

func (c *NacosClient) ListNacosServiceInstance(ctx context.Context, params map[string]string) (*ListNacosServiceInstanceRsp, error) {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_BASE_PATH+"/catalog/instances", params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: ListNacosServiceInstance error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: ListNacosServiceInstance success", zap.Any("param", params))
	var listNacosServiceInstanceRsp ListNacosServiceInstanceRsp
	err = json.Unmarshal([]byte(content), &listNacosServiceInstanceRsp)
	if err != nil {
		return nil, err
	}
	return &listNacosServiceInstanceRsp, nil
}

type NacosSubscriber struct {
	AddrStr     string `json:"addrStr"`
	Agent       string `json:"agent"`
	App         string `json:"app"`
	Ip          string `json:"ip"`
	Port        int32  `json:"port"`
	NamespaceId string `json:"namespaceId"`
	ServiceName string `json:"serviceName"`
	Cluster     string `json:"cluster"`
}

type SearchNacosSubscriberListRsp struct {
	NacosSubscribers []NacosSubscriber `json:"subscribers"`
	Count            int32             `json:"count"`
}

func (c *NacosClient) SearchNacosSubscriberList(ctx context.Context, params map[string]string) (*SearchNacosSubscriberListRsp, error) {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_INFO_PATH+"/subscribers", params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: SearchNacosServiceList error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: SearchHistoryNacosConfig success", zap.Any("param", params))
	var searchNacosSubscriberList SearchNacosSubscriberListRsp
	err = json.Unmarshal([]byte(content), &searchNacosSubscriberList)
	if err != nil {
		return nil, err
	}
	return &searchNacosSubscriberList, nil
}

func (c *NacosClient) CreateNacosService(ctx context.Context, params map[string]string) error {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_INFO_PATH, params, headers, http.MethodPost, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: CreateNacosService error", zap.Any("param", params), zap.Error(err))
		return err
	}
	log.Info(ctx, "Nacos: CreateNacosService success", zap.Any("param", params))
	if content != "ok" {
		return errors.New(content)
	}
	return nil
}

func (c *NacosClient) UpdateNacosService(ctx context.Context, params map[string]string) error {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_INFO_PATH, params, headers, http.MethodPut, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: CreateNacosService error", zap.Any("param", params), zap.Error(err))
		return err
	}
	log.Info(ctx, "Nacos: CreateNacosService success", zap.Any("param", params))
	if content != "ok" {
		return errors.New(content)
	}
	return nil
}

func (c *NacosClient) DeleteNacosService(ctx context.Context, params map[string]string) error {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.SERVICE_INFO_PATH, params, headers, http.MethodDelete, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: DeleteNacosService error", zap.Any("param", params), zap.Error(err))
		return err
	}
	log.Info(ctx, "Nacos: DeleteNacosService success", zap.Any("param", params))
	if content != "ok" {
		return errors.New(content)
	}
	return nil
}

type NacosConfigDetail struct {
	Id         string      `json:"id"`
	DataId     string      `json:"dataId"`
	Group      string      `json:"group"`
	Content    string      `json:"content"`
	Md5        string      `json:"md5"`
	Tenant     string      `json:"tenant"`
	AppName    string      `json:"appName"`
	Type       string      `json:"type"`
	CreateTime int64       `json:"createTime"`
	ModifyTime int64       `json:"modifyTime"`
	CreateUser interface{} `json:"createUser"`
	CreateIp   string      `json:"createIp"`
	Desc       interface{} `json:"desc"`
	Use        interface{} `json:"use"`
	Effect     interface{} `json:"effect"`
	Schema     interface{} `json:"schema"`
	ConfigTags interface{} `json:"configTags"`
}

func (c *NacosClient) GetNacosConfigDetail(ctx context.Context, params map[string]string) (*NacosConfigDetail, error) {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.CONFIG_PATH, params, headers, http.MethodGet, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: GetNacosConfigDetail error", zap.Any("param", params), zap.Error(err))
		return nil, err
	}
	log.Info(ctx, "Nacos: GetNacosConfigDetail success", zap.Any("param", params))
	var nacosConfigDetail NacosConfigDetail
	json.Unmarshal([]byte(content), &nacosConfigDetail)
	return &nacosConfigDetail, nil
}

func (c *NacosClient) EditNacosConfigDetail(ctx context.Context, params map[string]string, userName string) error {
	nacosConfigClient := c.configClient.(*config_client.ConfigClient)
	clientConfig, _ := nacosConfigClient.GetClientConfig()
	if len(clientConfig.NamespaceId) > 0 {
		params["tenant"] = clientConfig.NamespaceId
		params["namespaceId"] = clientConfig.NamespaceId
	}
	params["username"] = userName
	var headers = map[string]string{}
	headers["accessKey"] = clientConfig.AccessKey
	headers["secretKey"] = clientConfig.SecretKey
	content, err := c.nacosServer.ReqConfigApi(constant.CONFIG_PATH, params, headers, http.MethodPost, clientConfig.TimeoutMs)
	if err != nil {
		log.Error(ctx, "Nacos: EditNacosConfigDetail error", zap.Any("param", params), zap.Error(err))
		return err
	}
	log.Info(ctx, "Nacos: EditNacosConfigDetail success", zap.Any("param", params))
	if content != "true" {
		return errors.New(content)
	}
	return nil
}
