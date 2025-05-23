package grpc_client_interceptor

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log2 "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"panda-go/common/grpc-wrapper/grpc_client_wrapper/conn-storage"
	"panda-go/common/interceptor"
	_ "panda-go/common/interceptor"
	"panda-go/common/log"
	"panda-go/common/utils"
	"panda-go/common/utils/json-wrapper"
	"panda-go/component/k8s"
	"panda-go/component/nacos/config/grpc-call"
	"runtime/debug"
)

func JaegerGrpcUnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	// 出口处更新B3信息
	ctxNew, span := utils.StartSpanWrapper(ctx, cc.Target()+method, opentracing.Tag{Key: string(ext.Component), Value: "Grpc"})
	//span.SetBaggageItem()
	span.SetTag(interceptor.GRPC_TYPE, "client_unary")
	defer func() {
		if rErr := recover(); rErr != nil {
			log.Error(ctx, "[JaegerGrpcUnaryClientInterceptor] panic recover by Interceptor", zap.Any("rErr", rErr), zap.String("method", method), zap.Any("req", req), zap.ByteString("stack", debug.Stack()))
			err = status.Errorf(interceptor.PANIC_ERR_CODE, "%s", "server inner error")
			span.LogKV("Panic", utils.GetLessThenLimitStr(debug.Stack(), 800))
		}
		span.Finish()
	}()
	//span.LogKV("Param_Origin", req)
	span.LogFields(log2.String("Param", json_wrapper.MarshalToStringWithoutErr(req)))
	ctxNew = utils.SetContextWithB3(ctxNew, (span.Context()).(jaeger.SpanContext))
	err = invoker(ctxNew, method, req, reply, cc, opts...)
	if err != nil {
		// 处理错误
		ext.Error.Set(span, true)
		span.SetTag(interceptor.GRPC_STATUS_CODE, utils.ParseRPCErrorCode(err))
		log.Error(ctx, "JaegerGrpcUnaryClientInterceptor handler err", zap.Error(err))
		if len(err.Error()) > 200 {
			span.LogKV(interceptor.GRPC_ERROR_MSG, err.Error()[:800])
		} else {
			span.LogKV(interceptor.GRPC_ERROR_MSG, err)
		}
	} else {
		span.SetTag(interceptor.GRPC_STATUS_CODE, codes.OK)
	}
	return err
}

func JaegerGrpcStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (clientStream grpc.ClientStream, err error) {
	//ctxNew, err := utils.GetSpanCtxFromCtxB3KV(ctx)
	//if err != nil {
	//	println(err)
	//}

	clientStream, err = streamer(ctx, desc, cc, method, opts...)
	return
}

func RetryGrpcUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	retries := 3
	var isConnChange bool
	for i := 0; i < retries; i++ {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			if isConnChange {
				serviceName, _, err := utils.GetSvcNameAndNsFromTarget(cc.Target())
				if err == nil {
					conn_storage.SetServiceName2ConnMap(serviceName, cc)
				} else {
					log.Error(ctx, "[RetryGrpcUnaryClientInterceptor] utils.GetSvcNameAndNsFromTarget fail", zap.String("target", cc.Target()), zap.Error(err))
				}
			}
			return nil
		}
		code := utils.ParseRPCErrorCode(err)
		if code == int64(codes.Unavailable) {
			spanParent := opentracing.SpanFromContext(ctx)
			if spanParent != nil {
				ext.Error.Set(spanParent, true)
				spanParent.LogKV("RetryGrpcUnaryClientInterceptor fail", fmt.Sprintf("target = %s", cc.Target()))
			}
			log.Error(ctx, "Grpc call Unavailable", zap.String("target", cc.Target()), zap.String("method", method), zap.Any("req", req))
			serviceName, namespace, err := utils.GetSvcNameAndNsFromTarget(cc.Target())
			if err != nil {
				log.Error(ctx, "[RetryGrpcUnaryClientInterceptor] utils.GetSvcNameAndNsFromTarget failed", zap.String("method", method), zap.Any("req", req), zap.Error(err))
				return err
			}
			_, newNamespace, err := k8s.GetNewServiceEntity(ctx, serviceName, namespace)
			if err != nil {
				log.Error(ctx, fmt.Sprintf("[RetryGrpcUnaryClientInterceptor] try to get namespace of %s from k8s api failed", serviceName), zap.String("serviceName", serviceName))
				return err
			}
			grpc_call.GetGrpcCallConfig(ctx).SetServiceNamespace(ctx, serviceName, newNamespace)
			ccNew, err := grpc.Dial(utils.GetTarget(serviceName, newNamespace),
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
					JaegerGrpcUnaryClientInterceptor,
					RetryGrpcUnaryClientInterceptor)),
				grpc.WithStreamInterceptor(JaegerGrpcStreamClientInterceptor))
			if err != nil {
				log.Error(ctx, "[RetryGrpcUnaryClientInterceptor] grpc.Dial new target fail", zap.String("target", utils.GetTarget(serviceName, newNamespace)))
				return err
			} else {
				isConnChange = true
				cc = ccNew
				continue
			}
		}
		log.Infof(ctx, "Retry %d: %v\n", i+1, err)
	}
	return fmt.Errorf("Max retries reached")
}
