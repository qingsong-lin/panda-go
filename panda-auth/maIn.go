package main

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	traceLog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
	grpcMetadata "google.golang.org/grpc/metadata"
	"net"
	grpc_wrapper "panda-go/common/grpc-wrapper/grpc_server_wrapper"
	"panda-go/common/log"
	"panda-go/common/utils"
	"panda-go/proto/pb"
)

//func main() {
//	// 创建 Gin 引擎
//	r := gin.Default()
//
//	// 定义路由处理程序
//	r.GET("/test", handleTest)
//
//	// 启动 HTTP 服务器
//	if err := r.Run(":8080"); err != nil {
//		panic(err)
//	}
//}
//
//
//func handleTest(c *gin.Context) {
//	println("lgx")
//	tmp := 1
//	println(tmp)
//	//go func() {
//	//	for true {
//	//		time.Sleep(time.Second)
//	//		println("test hahahah lgx2")
//	//	}
//	//}()
//	c.JSON(http.StatusOK, gin.H{
//		"message": "Hello, World!",
//	})
//
//}

type HelloServer struct {
}

func (h *HelloServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	mdGrpc, okGrpc := grpcMetadata.FromIncomingContext(ctx)
	if okGrpc {
		for key, val := range mdGrpc {
			for _, v := range val {
				println(key, v)
			}
		}
	}
	ctxNew, span := utils.StartSpanWrapper(ctx, "ToServerCallSayHello", opentracing.Tag{Key: string(ext.Component), Value: "GrpcServerHandle"})
	log.Info(ctx, "HelloServer handle test begin!")
	//// 在 span 中添加一些标签或日志
	span.SetTag("custom-tag", "hahahahhahhahahahahha")
	span.LogFields(traceLog.String("event", "ToServerCallSayHello handling grpc request"))
	span.Finish()
	if cc, ok := pb.CreateClientWrapper[pb.PandaAccountClient](ctx).(pb.PandaAccountClient); ok {
		hello, err := cc.SayHello2DemoClient(ctxNew, &pb.HelloRequest{
			Name: "123231",
		})
		if err != nil {
			log.Error(ctx, "SayHello2DemoClient err", zap.Error(err))
			return &pb.HelloResponse{Reply: err.Error()}, nil
		}
		return &pb.HelloResponse{Reply: hello.GetReply()}, nil
	}
	return &pb.HelloResponse{Reply: "lgxsdjkl"}, nil
}

func NewHelloServer() *HelloServer {
	return &HelloServer{}
}

func main() {
	ctx := context.TODO()
	server := grpc_wrapper.NewServer()
	pb.RegisterPandaAuthServer(server, NewHelloServer())
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf(ctx, "failed to listen: %v", err)
	}
	log.Infof(ctx, "Server is listening on :50051")
	err = server.Serve(listener)
	if err != nil {
		log.Fatalf(ctx, "Failed to serve: %v", err)
	}
}
