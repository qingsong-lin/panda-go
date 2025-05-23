package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	grpc_wrapper "panda-go/common/grpc-wrapper/grpc_server_wrapper"
	"panda-go/common/grpc-wrapper/grpc_server_wrapper/prometheus_wrapper"
	"panda-go/common/log"
)

func main() {
	go func() {
		ctx := context.TODO()
		server := grpc_wrapper.NewServer()
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf(ctx, "failed to listen: %v", err)
		}
		log.Infof(ctx, "Server is listening on :50051")
		err = server.Serve(listener)
		if err != nil {
			log.Fatalf(ctx, "Failed to serve: %v", err)
		}
	}()
	// 注册 Goroutine 数量的指标
	prometheus_wrapper.InitCollectPrometheus()
	// 注册 /metrics 路由，Prometheus 将通过该路由拉取指标数据
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
