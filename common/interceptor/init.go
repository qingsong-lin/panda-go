package interceptor

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"panda-go/common/grpc-wrapper/grpc_server_wrapper/holms_wrapper"
	"panda-go/common/grpc-wrapper/grpc_server_wrapper/prometheus_wrapper"
	"panda-go/common/log"
	"panda-go/component/kafka/consumer"
	"panda-go/component/kafka/producer"
	"syscall"
)

const (
	GRPC_STATUS_CODE = "grpc.status_code"
	HTTP_STATUS_CODE = "http.status_code"
	GRPC_ERROR_MSG   = "Grpc_Err_Msg"
	PANIC_ERR_CODE   = 590
	GRPC_TYPE        = "grpc.type"
	HTTP_METHOD      = "http.method"
	HTTP_PROTO       = "http.proto"
)

func init() {
	ctx := context.TODO()
	cfg, err := config.FromEnv()
	if err != nil {
		log.Error(ctx, "无法解析  Jaeger 环境变量", zap.Any("err", err))
	}
	//cfg.Headers = &jaeger.HeadersConfig{
	//	TraceContextHeaderName: "x-b3-traceid",
	//}
	if cfg.ServiceName == "" {
		log.Error(ctx, "err no cfg.ServiceName")
		cfg.ServiceName = "panda-default"
	}
	// 初始化 Jaeger 追踪器
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatalf(ctx, "无法初始化 Jaeger 追踪器: %s", err.Error())
	}
	opentracing.SetGlobalTracer(tracer)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 创建一个 goroutine 来等待信号
	go func() {
		<-c
		closer.Close()
		holms_wrapper.StopHolmes()
		consumer.CloseKafkaClient()
		producer.CloseKafkaProducer()
		prometheus_wrapper.ClosePrometheus()
		log.Info(ctx, "accept quit signal, close global tracer!")
		// 执行清理操作的代码
		os.Exit(1)
	}()
}
