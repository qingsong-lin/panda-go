package pb

import (
	"context"
	"errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net/http"
	"panda-go/common/log"
)

func RegisterCommonServiceHandlerWithServiceName(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn, serviceName string) error {
	return RegisterCommonServiceHandlerClientWithServiceName(ctx, mux, NewCommonServiceClient(conn), serviceName)
}

func RegisterCommonServiceHandlerClientWithServiceName(ctx context.Context, mux *runtime.ServeMux, client CommonServiceClient, serviceName string) error {
	if serviceName == "" {
		return errors.New("[RegisterCommonServiceHandlerClientWithServiceName] serviceName is empty")
	}
	mux.HandlePath("GET", "/v1/"+serviceName+"/level/{level}", func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/pb.CommonService/SetLogLevel", runtime.WithHTTPPathPattern("/v1/"+serviceName+"/level"+"/{level}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_CommonService_SetLogLevel_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_CommonService_SetLogLevel_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})
	mux.HandlePath("POST", "/v1/"+serviceName+"/download-pprof", func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/pb.CommonService/DownloadPprofFile", runtime.WithHTTPPathPattern("/v1/download-pprof"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_CommonService_DownloadPprofFile_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		var fileChunk FileChunk
		err = resp.RecvMsg(&fileChunk)
		if err != nil {
			log.Error(ctx, "[RegisterCommonServiceHandlerClientWithServiceName] pprof down recvMsg from server ERR", zap.Error(err))
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		_, err = w.Write(fileChunk.Data)
		if err != nil {
			log.Error(ctx, "[RegisterCommonServiceHandlerClientWithServiceName] pprof down write to http client ERR", zap.Error(err))
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		for key, values := range md.HeaderMD {
			for _, value := range values {
				// 将 gRPC 的 metadata 写入 HTTP 头部
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	return nil
}
