package conn_storage

import (
	"errors"
	"google.golang.org/grpc"
	"sync"
)

var serviceName2ConnMap sync.Map

func SetServiceName2ConnMap(serviceName string, conn *grpc.ClientConn) {
	serviceName2ConnMap.Store(serviceName, conn)
}

func GetConnFromServiceMap(serviceName string) (*grpc.ClientConn, error) {
	connFromMap, ok := serviceName2ConnMap.Load(serviceName)
	if ok {
		cc, ok2 := connFromMap.(*grpc.ClientConn)
		if ok2 {
			return cc, nil
		}
	}
	return nil, errors.New("GetConnFromServiceMap get conn-storage from serviceName2ConnMap fail")
}

func DeleteByServiceKeys(serviceNames []string) {
	for _, serviceName := range serviceNames {
		serviceName2ConnMap.Delete(serviceName)
	}
}

func DeleteAllKey() {
	serviceName2ConnMap.Range(func(key, value any) bool {
		serviceName2ConnMap.Delete(key)
		return true
	})
}
