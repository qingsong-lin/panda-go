package main

import (
	"context"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"net/http"
	"panda-go/common/grpc-wrapper/grpc_client_wrapper"
	http_server_interceptor "panda-go/common/interceptor/http-server-interceptor"
	"panda-go/common/log"
	"panda-go/component/k8s"
	"panda-go/component/swagger"
	"panda-go/proto/pb"
	"path"
	"strings"
)

func serveSwaggerFile(w http.ResponseWriter, r *http.Request) {
	log.Info(context.TODO(), "start serveSwaggerFile")
	if !strings.HasSuffix(r.URL.Path, "swagger.json") {
		log.Info(context.TODO(), "Not Found: %s", zap.String("r.URL.Path", r.URL.Path))
		http.NotFound(w, r)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join("./proto/", p)
	log.Info(context.TODO(), fmt.Sprintf("Serving swagger-file: %s", p))
	http.ServeFile(w, r, p)
}

var (
	prefix2remove = []string{"/panda-gin-swagger", "/panda-grpc-gw"}
)

func RemovePrefix() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, prefix := range prefix2remove {
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, prefix, "", 1)
		}
	}
}

func serveSwaggerUI(mux *http.ServeMux) {
	fileServer := http.FileServer(&assetfs.AssetFS{
		Asset:    swagger.Asset,
		AssetDir: swagger.AssetDir,
		Prefix:   "third_party/swagger-ui",
	})
	prefix := "/swagger-ui/"
	mux.Handle(prefix, http.StripPrefix(prefix, fileServer))
}

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	c.Header("Access-Control-Allow-Credentials", "true")

	// 如果是预检请求，直接返回 200 状态码
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
		return
	}

	c.Next()
}

func serveK8sApi(w http.ResponseWriter, r *http.Request) {
	config, err := k8s.GetK8sConfig()
	if err != nil {
		log.Error(context.TODO(), "serveK8sApi rest.InClusterConfig failed ", zap.Error(err))
		http.Error(w, "Internal Server Error: serveK8sApi rest.InClusterConfig failed", http.StatusInternalServerError)
		return
	}
	t, err := rest.TransportFor(config)
	url := *r.URL
	url.Host = config.Host
	url.Scheme = "http"
	httpProxy := proxy.NewUpgradeAwareHandler(&url, t, true, false, nil)
	httpProxy.ServeHTTP(w, r)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gwMux := runtime.NewServeMux()
	conn, err := grpc_client_wrapper.DialWrapper(ctx, "panda-account")
	if err == nil {
		pb.RegisterPandaAccountHandler(ctx, gwMux, conn)
		pb.RegisterCommonServiceHandlerWithServiceName(ctx, gwMux, conn, "panda-account")
	} else {
		log.Error(ctx, "DialWrapper <panda-account> fail", zap.Error(err))
	}
	conn, err = grpc_client_wrapper.DialWrapper(ctx, "panda-auth")
	if err == nil {
		pb.RegisterPandaAuthHandler(ctx, gwMux, conn)
		pb.RegisterCommonServiceHandlerWithServiceName(ctx, gwMux, conn, "panda-auth")
	} else {
		log.Error(ctx, "DialWrapper <panda-auth> fail", zap.Error(err))
	}
	conn, err = grpc_client_wrapper.DialWrapper(ctx, "panda-export")
	if err == nil {
		pb.RegisterCommonServiceHandlerWithServiceName(ctx, gwMux, conn, "panda-export")
	} else {
		log.Error(ctx, "DialWrapper <panda-auth> fail", zap.Error(err))
	}
	//gwMux.HandlePath()
	// 创建 Gin 引擎
	engine := gin.Default()
	engine.Use(RemovePrefix())
	engine.Use(http_server_interceptor.JaegerHTTPServerInterceptor())
	//gwGroup := engine.Group("/panda-grpc-gw")
	//gwGroup.Use(RemovePrefix())
	//gwGroup.Any("/*any", gin.WrapH(gwMux))
	//swaggerGroup := engine.Group("/panda-gin-swagger")
	//swaggerGroup.Use(RemovePrefix())
	mux := http.NewServeMux()
	mux.Handle("/", gwMux)
	mux.HandleFunc("/swagger/", serveSwaggerFile)
	mux.HandleFunc("/k8s/", serveK8sApi)
	serveSwaggerUI(mux)
	//swaggerGroup.Use(corsMiddleware)
	engine.Any("/*any", gin.WrapH(mux))
	log.Info(ctx, "start client gin")
	// 启动 GIN 服务器
	if err := engine.Run(":8080"); err != nil {
		panic(err)
	}
}
