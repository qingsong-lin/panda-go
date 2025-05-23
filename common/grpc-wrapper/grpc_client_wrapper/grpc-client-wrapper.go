package grpc_client_wrapper

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"panda-go/common/grpc-wrapper/grpc_client_wrapper/conn-storage"
	"panda-go/common/interceptor/grpc-client-interceptor"
	"panda-go/common/log"
	"panda-go/common/utils"
	"panda-go/component/k8s"
	"panda-go/component/nacos/config/grpc-call"
)

func DialWrapper(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
	namespace := grpc_call.GetGrpcCallConfig(ctx).GetServiceNamespace(serviceName)
	if namespace == "" {
		_, newNamespace, err := k8s.GetNewServiceEntity(ctx, serviceName, namespace)
		if err != nil {
			log.Error(ctx, fmt.Sprintf("[DialWrapper] try to get namespace of %s from k8s api failed", serviceName), zap.String("serviceName", serviceName))
			return nil, err
		}
		namespace = newNamespace
		grpc_call.GetGrpcCallConfig(ctx).SetServiceNamespace(ctx, serviceName, namespace)
	}
	return grpc.Dial(utils.GetTarget(serviceName, namespace),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_client_interceptor.JaegerGrpcUnaryClientInterceptor,
			grpc_client_interceptor.RetryGrpcUnaryClientInterceptor)),
		grpc.WithStreamInterceptor(grpc_client_interceptor.JaegerGrpcStreamClientInterceptor),
	)
}

func GetConn(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
	myConn, err := conn_storage.GetConnFromServiceMap(serviceName)
	if err == nil {
		return myConn, nil
	} else {
		log.Warn(ctx, "[GetConn] fail", zap.Error(err))
	}
	cc, err := DialWrapper(ctx, serviceName)
	if err != nil {
		log.Error(ctx, "err")
		return nil, err
	}
	if myConn, _ = conn_storage.GetConnFromServiceMap(serviceName); myConn == nil {
		conn_storage.SetServiceName2ConnMap(serviceName, cc)
	}
	return cc, nil
}
