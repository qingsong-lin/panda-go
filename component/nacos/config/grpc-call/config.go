package grpc_call

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"os"
	conn_storage "panda-go/common/grpc-wrapper/grpc_client_wrapper/conn-storage"
	"panda-go/common/log"
	"panda-go/common/utils"
	json_wrapper "panda-go/common/utils/json-wrapper"
	panda_nacos "panda-go/component/nacos/base"
	"sync"
)

var (
	once                sync.Once
	GrpcCallRelationKey = ".grpc.call.map.service2namespace"
	config              *grpcCallConfig
)

type grpcCallConfig struct {
	service2NamespaceMap sync.Map
}

func GetGrpcCallConfig(ctx context.Context) *grpcCallConfig {
	once.Do(func() {
		serviceName := os.Getenv(panda_nacos.SERVICE_NAME)
		if serviceName == "" {
			serviceName = "common-grpc-relation"
		}
		GrpcCallRelationKey = serviceName + GrpcCallRelationKey
		config = &grpcCallConfig{}
		panda_nacos.GetNacosClient().GetAndWatchConfig(ctx, GrpcCallRelationKey, func(key, value string) {
			tmpMap := make(map[string]string)
			ctxNew := context.TODO()
			if value != "" {
				err := json.Unmarshal([]byte(value), &tmpMap)
				if nil != err {
					log.Error(ctxNew, "Unmarshal fail ", zap.String("value", value))
				}
				oldMap := utils.GetNormalMapFromSyncMap(&config.service2NamespaceMap)
				diffMap := make(map[string]string)
				var diffKeys []string
				for svcName, ns := range tmpMap {
					if oldNs, ok := oldMap[svcName]; ok && ns == oldNs {
						continue
					}
					diffMap[svcName] = ns
					diffKeys = append(diffKeys, svcName)
				}
				if len(diffMap) != 0 {
					log.Info(ctxNew, "GetGrpcCallConfig change", zap.Any("diffMap", diffMap))
				}

				conn_storage.DeleteByServiceKeys(diffKeys)
				utils.StorSyncMapFromNormalMap(diffMap, &config.service2NamespaceMap)
			}
		})
	})
	return config
}

func (c *grpcCallConfig) GetServiceNamespace(serviceName string) string {
	if namespace, ok := c.service2NamespaceMap.Load(serviceName); ok {
		if ns, ok2 := namespace.(string); ok2 {
			return ns
		}
	}
	return ""
}

func (c *grpcCallConfig) SetServiceNamespace(ctx context.Context, serviceName string, nameSpace string) {
	c.service2NamespaceMap.Store(serviceName, nameSpace)
	mp := utils.GetNormalMapFromSyncMap(&c.service2NamespaceMap)
	if len(mp) != 0 {
		_, err := panda_nacos.GetNacosClient().PublishConfig(ctx, &panda_nacos.ConfigParam{
			DataId:  GrpcCallRelationKey,
			Content: json_wrapper.MarshalToStringWithoutErr(mp),
		}, "")
		if err != nil {
			log.Error(ctx, "[grpcCallConfig] SetServiceNamespace nacos PublishConfig err", zap.Error(err), zap.String("serviceName", serviceName), zap.String("nameSpace", nameSpace))
			return
		}
	} else {
		log.Error(ctx, "[grpcCallConfig] SetServiceNamespace parse from syncMap empty")
	}
	return
}
