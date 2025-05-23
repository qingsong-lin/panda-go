package http_server_interceptor

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log2 "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"net/http"
	"panda-go/common/interceptor"
	"panda-go/common/log"
	"panda-go/common/utils"
	"runtime/debug"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) WriteString(s string) (n int, err error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}
func (r responseBodyWriter) Write(s []byte) (n int, err error) {
	r.body.Write(s)
	return r.ResponseWriter.Write(s)
}

func JaegerHTTPServerInterceptor() (handler gin.HandlerFunc) {
	// http传递因为client无法发送ctx内容，只能发送header，在服务端处由header生成相关的ctx
	// 所以暂时header只维护B3头的包装和解析，不维护uber-trace-id（上游发出如有uber-trace-id也转成B3头）
	handler = func(c *gin.Context) {
		ctx := c.Request.Context()
		header := c.Request.Header
		log.Debug(ctx, "HTTPServerInterceptor", zap.Any("header", header), zap.String("URI", c.Request.RequestURI))
		ctx = utils.UserHeaderToSetContextWithB3(ctx, header)
		spanCtx, _ := utils.GetSpanCtxFromCtxB3KV(ctx)
		tracer := opentracing.GlobalTracer()
		span := tracer.StartSpan(c.Request.RequestURI, opentracing.Tag{Key: string(ext.Component), Value: "Http"}, ext.RPCServerOption(spanCtx))
		span.SetTag(interceptor.HTTP_METHOD, c.Request.Method)
		span.SetTag(interceptor.HTTP_PROTO, c.Request.Proto)
		//span.SetBaggageItem()
		defer func() {
			if rErr := recover(); rErr != nil {
				log.Error(ctx, "[JaegerHTTPServerInterceptor] panic recover by Interceptor", zap.Any("rErr", rErr), zap.ByteString("stack", debug.Stack()))
				ext.Error.Set(span, true)
				panicStr := utils.GetLessThenLimitStr(debug.Stack(), 800)
				span.LogKV("Panic", panicStr)
				c.JSON(interceptor.PANIC_ERR_CODE, panicStr)
			}
			if c.Writer.Status() != http.StatusOK {
				ext.Error.Set(span, true)
				if w, ok := c.Writer.(*responseBodyWriter); ok {
					span.LogFields(log2.Object("err_msg", w.body.String()))
				}
			}
			span.SetTag(interceptor.HTTP_STATUS_CODE, c.Writer.Status())
			span.Finish()
		}()
		KVParam := utils.GetRequestKVParamStr(c.Request)
		if KVParam != "" {
			span.LogFields(log2.String("KV_Param", KVParam))
		}
		BodyParam := utils.GetRequestBodyParamStr(c.Request)
		if BodyParam != "" {
			span.LogFields(log2.String("Body_Param", BodyParam))
		}
		ctxNew := opentracing.ContextWithSpan(ctx, span)
		ctxNew = utils.SetContextWithB3(ctxNew, (span.Context()).(jaeger.SpanContext))
		c.Request = c.Request.WithContext(ctxNew)
		c.Next()
	}
	return
}
