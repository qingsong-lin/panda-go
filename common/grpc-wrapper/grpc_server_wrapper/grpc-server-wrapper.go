package grpc_server_wrapper

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"os"
	"panda-go/common/grpc-wrapper/grpc_server_wrapper/holms_wrapper"
	"panda-go/common/grpc-wrapper/grpc_server_wrapper/prometheus_wrapper"
	"panda-go/common/interceptor/grpc-server-interceptor"
	"panda-go/common/log"
	"panda-go/proto/pb"
	"runtime/pprof"
	"time"
)

type GrpcCommonServer struct {
}

func (c *GrpcCommonServer) SetLogLevel(ctx context.Context, in *pb.SetLogLevelRequest) (*pb.SetLogLevelResponse, error) {
	err := log.SetLevel(in.Level)
	if err != nil {
		return &pb.SetLogLevelResponse{
			Code:   400,
			ErrMsg: err.Error(),
		}, err
	}
	return &pb.SetLogLevelResponse{
		Code: 200,
	}, nil
}

func (c *GrpcCommonServer) DownloadPprofFile(req *pb.DownloadPprofFileRequest, ss pb.CommonService_DownloadPprofFileServer) error {
	log.Info(ss.Context(), "DownloadPprofFile start")
	if req == nil {
		return errors.New(" req nil!")
	}
	var selectPprofTypes []string
	if len(req.PprofTypes) == 0 {
		for key := range pb.DownloadPprofFileRequest_PprofType_value {
			selectPprofTypes = append(selectPprofTypes, key)
		}
	} else {
		for _, pprofType := range req.GetPprofTypes() {
			selectPprofTypes = append(selectPprofTypes, pprofType.String())
		}
	}
	// 创建一个内存中的ZIP压缩文件
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)
	for _, pprofType := range selectPprofTypes {
		file, err := zipWriter.Create(pprofType)
		if err != nil {
			log.Error(ss.Context(), "[DownloadPprofFile] zipWriter.Create fail", zap.Error(err))
			return err
		}
		_ = pprof.Lookup(pprofType).WriteTo(file, 0)
	}
	zipWriter.Close()
	// 设置响应头，告诉浏览器这是一个ZIP文件下载
	ss.SetHeader(map[string][]string{
		"Content-Disposition": {"attachment; filename=pprof-files.zip"},
		"Content-Type":        {"application/zip"},
		"Content-Length":      {string(zipBuffer.Len())},
	})
	content := zipBuffer.Bytes()
	if len(content) == 0 {
		log.Error(ss.Context(), "[DownloadPprofFile] zipBuffer.Bytes() is empty")
		return errors.New("[DownloadPprofFile] zipBuffer.Bytes() is empty")
	}
	err := os.WriteFile("/tmp/pprof.zip", content, 0644)
	if err != nil {
		log.Error(ss.Context(), "[DownloadPprofFile] Error writing zip file", zap.Error(err))
	}
	ss.Send(&pb.FileChunk{
		Data: content,
	})
	log.Info(ss.Context(), "DownloadPprofFile end")
	return nil
}

func NewServer() *grpc.Server {
	holms_wrapper.StartHolmes()
	prometheus_wrapper.InitPushPrometheus()
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_server_interceptor.JaegerGrpcUnaryServerInterceptor,
		)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 180 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.StreamInterceptor(grpc_server_interceptor.JaegerGrpcStreamServerInterceptor),
	)
	pb.RegisterCommonServiceServer(server, &GrpcCommonServer{})
	return server
}
