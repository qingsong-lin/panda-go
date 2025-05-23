package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	traceLog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	grpc_wrapper "panda-go/common/grpc-wrapper/grpc_server_wrapper"
	"panda-go/common/interceptor/http-server-interceptor"
	"panda-go/common/log"
	"panda-go/common/utils"
	"panda-go/panda-account/server"
	"panda-go/proto/pb"
)

func main() {
	ctx := context.TODO()
	go func() {
		grpcServer := grpc_wrapper.NewServer()
		pb.RegisterPandaAccountServer(grpcServer, server.NewPandaAccountServer())
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf(ctx, "failed to listen: %v", err)
		}
		log.Info(ctx, "panada panda-go Server is listening on :50051")
		err = grpcServer.Serve(listener)
		if err != nil {
			log.Fatalf(ctx, "Failed to serve: %v", err)
		}
	}()
	// 创建 Gin 引擎
	r := gin.Default()
	r.Use(http_server_interceptor.JaegerHTTPServerInterceptor())
	// 定义路由处理程序
	r.GET("/test", handleTest)
	r.POST("/test", handlerPostTest)
	log.Info(ctx, "start client gin")
	// 启动 HTTP 服务器
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func handlerPostTest(c *gin.Context) {
	defer func() {
		c.Request.Body.Close()
	}()
	content, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"format":  c.Request.PostForm,
	})
}

func handleTest(c *gin.Context) {
	ctx := c.Request.Context()
	ctxNew, span := utils.StartSpanWrapper(ctx, "handleGetTest", opentracing.Tag{Key: string(ext.Component), Value: "HTTPHandler"})
	//// 在 span 中添加一些标签或日志
	span.SetTag("custom-tag", "hahahahhahhahahahahha")
	span.LogFields(traceLog.String("event", "handling HTTP request"))
	span.Finish()
	//println("incomming")
	//mdGrpc, okGrpc := grpcMetadata.FromIncomingContext(ctx)
	//if okGrpc {
	//	for key, val := range mdGrpc {
	//		for _, v := range val {
	//			println(key, v)
	//		}
	//
	//	}
	//}
	//println("outcomming")
	//mdGrpc, okGrpc = grpcMetadata.FromOutgoingContext(ctx)
	//if okGrpc {
	//	for key, val := range mdGrpc {
	//		for _, v := range val {
	//			println(key, v)
	//		}
	//
	//	}
	//}
	//md := metautils.ExtractIncoming(ctx)
	//if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
	//println("parantspan not nil")
	//span1, ctx1 := opentracing.StartSpanFromContext(ctx, "panda-go-client-read-redis", opentracing.Tag{Key: string(ext.Component), Value: "preStart"}, opentracing.Tag{Key: "func", Value: "handleTest"})
	//span1.LogFields(traceLog.String("event", fmt.Sprintf("start CorsFilter")))
	//span2, ctx2 := opentracing.StartSpanFromContext(ctx1, "panda-go-client-database", opentracing.Tag{Key: string(ext.Component), Value: "mysql"}, opentracing.Tag{Key: "func", Value: "delete database item"})
	//span2.LogFields(traceLog.String("request", fmt.Sprintf("{\"key\":\"value\"}")))
	//span2.Finish()
	//span1.Finish()
	if cc, ok := pb.CreateClientWrapper[pb.PandaAuthClient](ctx).(pb.PandaAuthClient); ok {
		hello, err := cc.SayHello(ctxNew, &pb.HelloRequest{
			Name: "11",
		})
		if err != nil {
			log.Error(ctx, "SayHello err", zap.Error(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": hello,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}
