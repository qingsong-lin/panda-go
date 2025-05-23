package grpc_server_interceptor

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log2 "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"panda-go/common/interceptor"
	_ "panda-go/common/interceptor"
	"panda-go/common/log"
	"panda-go/common/utils"
	"panda-go/common/utils/json-wrapper"
	"runtime/debug"
)

func JaegerGrpcUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// 先查看B3头存在不存在，不存在可能是上游就没有经历过istio envoy或者就是本节点就是链头，打印一条grpc span继续往下走
	// B3存在，有可能是经历过istio envoy，或者是上游输出outContext时，更新了B3信息
	spanCtx, _ := utils.GetSpanCtxFromCtxB3KV(ctx)
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan(info.FullMethod, opentracing.Tag{Key: string(ext.Component), Value: "Grpc"}, ext.RPCServerOption(spanCtx))
	//span.SetBaggageItem()
	defer func() {
		if rErr := recover(); rErr != nil {
			log.Error(ctx, "[JaegerGrpcUnaryServerInterceptor] panic recover by Interceptor", zap.Any("rErr", rErr), zap.String("info.FullMethod", info.FullMethod), zap.Any("req", req), zap.ByteString("stack", debug.Stack()))
			err = status.Errorf(interceptor.PANIC_ERR_CODE, "%s", "server inner error")
			ext.Error.Set(span, true)
			span.LogKV("Panic", utils.GetLessThenLimitStr(debug.Stack(), 800))
		}
		span.Finish()
	}()
	//span.LogKV("Param_Origin", req)
	span.LogFields(log2.String("Param", json_wrapper.MarshalToStringWithoutErr(req)))
	span.SetTag(interceptor.GRPC_TYPE, "server_unary")
	ctxNew := opentracing.ContextWithSpan(ctx, span)
	resp, err = handler(ctxNew, req)
	if err != nil {
		// 处理错误
		ext.Error.Set(span, true)
		span.SetTag(interceptor.GRPC_STATUS_CODE, utils.ParseRPCErrorCode(err))
		log.Error(ctx, "JaegerGrpcUnaryServerInterceptor handler err", zap.Error(err))
		if len(err.Error()) > 200 {
			span.LogKV(interceptor.GRPC_ERROR_MSG, err.Error()[:800])
		} else {
			span.LogKV(interceptor.GRPC_ERROR_MSG, err)
		}
	} else {
		span.SetTag(interceptor.GRPC_STATUS_CODE, codes.OK)
	}
	return
}

func JaegerGrpcStreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	spanCtx, _ := utils.GetSpanCtxFromCtxB3KV(ss.Context())
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan(info.FullMethod, opentracing.Tag{Key: string(ext.Component), Value: "Grpc"}, ext.RPCServerOption(spanCtx))
	//span.SetBaggageItem()
	defer func() {
		if rErr := recover(); rErr != nil {
			log.Error(ss.Context(), "[JaegerGrpcStreamServerInterceptor] panic recover by Interceptor", zap.Any("rErr", rErr), zap.String("info.FullMethod", info.FullMethod), zap.ByteString("stack", debug.Stack()))
			err = status.Errorf(interceptor.PANIC_ERR_CODE, "%s", "server inner error")
			span.LogKV("Panic", utils.GetLessThenLimitStr(debug.Stack(), 800))
		}
		span.Finish()
	}()
	span.SetTag(interceptor.GRPC_TYPE, "server_stream")
	md, _ := metadata.FromIncomingContext(ss.Context())
	md = utils.SetMdWithB3(md, span.Context())
	ss.SetHeader(md)
	ss.SetTrailer(md)
	err = handler(srv, ss)
	if err != nil {
		// 处理错误
		log.Error(ss.Context(), "JaegerGrpcStreamServerInterceptor handler err", zap.Error(err))
		span.SetTag(interceptor.GRPC_STATUS_CODE, utils.ParseRPCErrorCode(err))
		if len(err.Error()) > 200 {
			span.LogKV(interceptor.GRPC_ERROR_MSG, err.Error()[:800])
		} else {
			span.LogKV(interceptor.GRPC_ERROR_MSG, err)
		}
	} else {
		span.SetTag(interceptor.GRPC_STATUS_CODE, codes.OK)
	}
	return
}
