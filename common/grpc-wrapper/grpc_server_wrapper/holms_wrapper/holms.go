package holms_wrapper

import (
	"context"
	"encoding/json"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"go.uber.org/zap"
	"mosn.io/holmes"
	mlog "mosn.io/pkg/log"
	"panda-go/common/log"
	json_wrapper "panda-go/common/utils/json-wrapper"
	panda_nacos "panda-go/component/nacos/base"
	"runtime"
	"time"
)

var s *holmsManager

func init() {
	s = &holmsManager{}
}

func StartHolmes() {
	go s.init()
}

func StopHolmes() {
	if s != nil {
		s.stopTriggerHolmes(context.TODO())
	}
}

type ItemConfig struct {
	Type     string `json:"type"`
	Min      int    `json:"min"`
	Diff     int    `json:"diff"`
	Abs      int    `json:"abs"`
	Max      int    `json:"max"`
	CoolDown int32  `json:"cool_down"` // 单位秒
}

type RealTimeConfig struct {
	AlertName  string `json:"alert_name"`
	Type       string `json:"type"` //cpu/memory/goroutine
	Pod        string `json:"pod"`
	Seconds    string `json:"seconds"`
	Debug      string `json:"debug"`
	CreateTime int64  `json:"create_time"` //超过2min则认为配置已经失效
}

type HolmesConfig struct {
	Enable bool          `json:"enable"`
	Items  []*ItemConfig `json:"items"`

	Trigger struct {
		Name string `json:"name"`
	} `json:"trigger"`
}
type holmsManager struct {
	triggerHolmes *holmes.Holmes
	//realTimeHolmes *holmes.Holmes

	holmesConfig HolmesConfig
}

func (m *holmsManager) stopTriggerHolmes(ctx context.Context) {
	if m.triggerHolmes != nil {
		m.triggerHolmes.Stop()
		// 关闭对缩的跟踪
		runtime.SetMutexProfileFraction(0)
		// 关闭对阻塞操作的跟踪
		runtime.SetBlockProfileRate(0)
		runtime.SetCPUProfileRate(0)
		m.triggerHolmes = nil
		log.Info(ctx, "[stopTriggerHolmes] stop trigger holmes")
	}
}
func (m *holmsManager) init() {
	ctx := context.TODO()
	key := panda_nacos.ServerName + ".holmes.config"
	handler := func(key, val string) {
		log.Info(ctx, "holmes handler", zap.String("key", key), zap.String("value", val))
		m.stopTriggerHolmes(ctx)
		if len(val) == 0 {
			return
		}
		var data *HolmesConfig
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			logger.Error("marshal config fail", zap.Error(err))
			return
		}
		m.holmesConfig = *data
		m.createTriggerHolmes(ctx)
	}
	var holmesConfig *HolmesConfig
	data, err := panda_nacos.GetNacosClient().GetConfig(ctx, &panda_nacos.ConfigParam{
		DataId: key,
	})
	if err != nil {
		// TODO:区分是不存在还是连接报错
		holmesConfig = &HolmesConfig{
			Enable: true,
			Items: []*ItemConfig{
				{
					Type:     "goroutine",
					Min:      1000,
					Diff:     25,
					Abs:      20 * 1000,
					Max:      100 * 1000,
					CoolDown: 2 * 60,
				},
				{
					Type:     "cpu",
					Min:      10,
					Diff:     25,
					Abs:      80,
					Max:      80,
					CoolDown: 2 * 60,
				},
				{
					Type:     "allocs",
					Min:      40,
					Diff:     25,
					Abs:      80,
					Max:      80,
					CoolDown: 2 * 60,
				},
				{
					Type:     "heap",
					Min:      10,
					Diff:     20,
					Abs:      40,
					CoolDown: 2 * 60,
				},
			},
			Trigger: struct {
				Name string `json:"name"`
			}{
				Name: panda_nacos.ServerName,
			},
		}
		data = json_wrapper.MarshalToStringWithoutErr(holmesConfig)
		_, err := panda_nacos.GetNacosClient().PublishConfig(ctx, &panda_nacos.ConfigParam{
			DataId:  key,
			Content: data,
		}, "")
		if err != nil {
			log.Error(ctx, "[init] holmsManager PublishConfig failed", zap.Error(err))
			return
		}
	}
	handler(key, data)
	panda_nacos.GetNacosClient().WatchConfig(ctx, &panda_nacos.ConfigParam{
		DataId: key,
	}, handler)
	runtime.SetMutexProfileFraction(1)
	//// 开启对阻塞操作的跟踪
	runtime.SetBlockProfileRate(1)
	log.Info(ctx, "[init] holmsManager init end")
}

func (m *holmsManager) createTriggerHolmes(ctx context.Context) {
	if m.holmesConfig.Enable == false {
		return
	}
	// 开启对缩的跟踪
	runtime.SetMutexProfileFraction(1)
	// 开启对阻塞操作的跟踪
	runtime.SetBlockProfileRate(1)
	options := make([]holmes.Option, 0)
	for _, item := range m.holmesConfig.Items {
		switch item.Type {
		case "cpu":
			options = append(options, holmes.WithCPUDump(item.Min, item.Diff, item.Abs, time.Duration(item.CoolDown)*time.Second))
			options = append(options, holmes.WithCPUMax(item.Max))
		case "goroutine":
			options = append(options, holmes.WithGoroutineDump(item.Min, item.Diff, item.Abs, item.Max, time.Duration(item.CoolDown)*time.Second))
		case "allocs":
			options = append(options, holmes.WithMemDump(item.Min, item.Diff, item.Abs, time.Duration(item.CoolDown)*time.Second))
		case "thread":
			options = append(options, holmes.WithThreadDump(item.Min, item.Diff, item.Abs, time.Duration(item.CoolDown)*time.Second))
		case "heap":
			options = append(options, holmes.WithGCHeapDump(item.Min, item.Diff, item.Abs, time.Duration(item.CoolDown)*time.Second))
		}
	}
	if len(options) == 0 {
		log.Info(ctx, "[createTriggerHolmes] options is empty", zap.Any("items", m.holmesConfig.Items))
	}
	if handler, err := m.newHolmes("trigger", m.holmesConfig.Trigger.Name, options); err != nil {
		logger.Error("create trigger holmes fail", zap.Error(err))
	} else {
		m.triggerHolmes = handler
		for _, item := range m.holmesConfig.Items {
			switch item.Type {
			case "cpu":
				m.triggerHolmes.EnableCPUDump()
			case "goroutine":
				m.triggerHolmes.EnableGoroutineDump()
			case "allocs":
				m.triggerHolmes.EnableMemDump()
			case "thread":
				m.triggerHolmes.EnableThreadDump()
			case "heap":
				m.triggerHolmes.EnableGCHeapDump()
			}
		}
		m.triggerHolmes.Start()
		logger.Info("start trigger holmes")
	}
}

func (m *holmsManager) newHolmes(appName string, alertName string, opts []holmes.Option) (*holmes.Holmes, error) {
	logger := holmes.NewStdLogger()
	logger.SetLogLevel(mlog.WARN)

	r := &ReporterImpl{app: appName, alertName: alertName}
	opts = append([]holmes.Option{
		holmes.WithProfileReporter(r),

		holmes.WithDumpPath("/tmp/profile"),
		// 日志等级 warn，输出到标准输出
		holmes.WithLogger(logger),

		// 在 Docker 或者 CGroup 模式下使用
		// Note: 目前测试在 Docker 开启有问题
		//holmes.WithCGroup(runtime.GOOS == "linux"),
	},
		// 用户自定义配置会覆盖默认配置
		opts...)

	tmp, err := holmes.New(opts...)
	return tmp, err
}

type ReporterImpl struct {
	app       string
	alertName string
}

type HolmesData struct {
	Type      string `json:"type"` //realtime / trigger
	AlertName string `json:"alert_name"`
	Resource  string `json:"resource"`
	EndTime   int64  `json:"end_time"`
	Content   string `json:"content"`
	Detail    string `json:"detail"`
}

func (r *ReporterImpl) Report(pType string, buf []byte, reason string, eventID string) error {
	logger.Info("holmes report", zap.String("pType", pType),
		zap.String("reason", reason), zap.String("eventID", eventID), zap.Int("buf_size", len(buf)))
	return nil
}
