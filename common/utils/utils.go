package utils

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"github.com/uber/jaeger-client-go"
	"google.golang.org/grpc/codes"
	grpcMetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// CopyContext copy incoming and outgoing metadata, and trace span into an empty context
func CopyContext(ctx context.Context) context.Context {

	//newCtx := opentracing.ContextWithSpan(context.TODO(), opentracing.SpanFromContext(ctx))
	var newCtx context.Context
	var mdGrpc grpcMetadata.MD
	mdGrpc, okGrpc := grpcMetadata.FromIncomingContext(ctx)
	if okGrpc {
		newCtx = grpcMetadata.NewIncomingContext(newCtx, mdGrpc)
	}
	mdGrpc, okGrpc = grpcMetadata.FromOutgoingContext(ctx)
	if okGrpc {
		newCtx = grpcMetadata.NewOutgoingContext(newCtx, mdGrpc)
	}

	return newCtx
}

func InAndOutContextAppendKV(ctx context.Context, k, v string) context.Context {
	md, _ := grpcMetadata.FromIncomingContext(ctx)
	md = grpcMetadata.Join(md, grpcMetadata.Pairs(k, v))
	ctx = grpcMetadata.NewIncomingContext(ctx, md)
	return grpcMetadata.AppendToOutgoingContext(ctx, k, v)
}

func InAndOutContextSetKV(ctx context.Context, k, v string) context.Context {
	mdIn, okIn := grpcMetadata.FromIncomingContext(ctx)
	if !okIn {
		mdIn = grpcMetadata.MD{}
	}
	mdIn.Set(k, v)
	mdOut, okOut := grpcMetadata.FromOutgoingContext(ctx)
	if !okOut {
		mdOut = grpcMetadata.MD{}
	}
	mdOut.Set(k, v)
	ctx = grpcMetadata.NewIncomingContext(ctx, mdIn)
	return grpcMetadata.NewOutgoingContext(ctx, mdOut)
}

func InComing2OutComing(ctx context.Context) context.Context {
	md, _ := grpcMetadata.FromIncomingContext(ctx)
	return grpcMetadata.NewOutgoingContext(ctx, md)
}

func GetB3CombineTraceIdAndUberTraceId(mp map[string][]string) (B3CombineTraceId, UberTraceId string) {
	var traceId, spanId, parentSpanID, flag string
	for key, values := range mp {
		//if utils.CmpWithStringListAnyEqualFold(key, []string{"x-b3-traceid", "x-b3-spanid"}) {
		//	ctx = utils.InAndOutContextAppendKV(ctx, key, c.GetHeader(key))
		//}
		if len(values) == 0 {
			continue
		}
		if strings.EqualFold(key, "x-b3-traceid") {
			traceId = values[0]
		} else if strings.EqualFold(key, "x-b3-spanid") {
			spanId = values[0]
		} else if strings.EqualFold(key, "x-b3-parentspanid") {
			parentSpanID = values[0]
		} else if strings.EqualFold(key, "x-b3-sampled") {
			flag = values[0]
		} else if strings.EqualFold(key, "uber-trace-id") {
			UberTraceId = values[0]
		}
	}
	if traceId != "" {
		B3CombineTraceId = B3TraceIdToUberTraceId(traceId, spanId, parentSpanID, flag)
	}
	return
}

// StartSpanWrapper  服务内部添加多个span时候，未跨grpc服务时候使用该函数
func StartSpanWrapper(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (reCtx context.Context, reSpan opentracing.Span) {
	// 先从ctx里面拿spanContext，如果拿得到则从spanContext转化链路,
	// 如果拿不到，则从ctx里拿b3继续链路
	spanParent := opentracing.SpanFromContext(ctx)
	if spanParent != nil {
		optNew := append(opts, opentracing.ChildOf(spanParent.Context()))
		reSpan = spanParent.Tracer().StartSpan(operationName, optNew...)
	} else {
		md, ok := grpcMetadata.FromIncomingContext(ctx)
		if !ok {
			md = grpcMetadata.MD{}
		}
		B3CombineTraceId, _ := GetB3CombineTraceIdAndUberTraceId(md)
		if B3CombineTraceId != "" {
			md["uber-trace-id"] = []string{B3CombineTraceId}
		}
		tracer := opentracing.GlobalTracer()
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md))
		// 创建 span
		opts = append(opts, ext.RPCServerOption(spanCtx))
		reSpan = tracer.StartSpan(operationName, opts...)
	}

	// 在 gRPC 请求中传递 span context
	reCtx = opentracing.ContextWithSpan(ctx, reSpan)
	//carrier := make(map[string]string)
	//// 使用 Inject 方法将追踪信息注入到 carrier 中
	//err := span.Tracer().Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(carrier))
	//if err != nil {
	//	log.Error(ctx, zap.Error(err))
	//}
	//for key, val := range carrier {
	//	ctx = utils.InAndOutContextSetKV(ctx, key, val)
	//}
	return
}

func CmpWithStringListAnyEqualFold(key string, in []string) bool {
	for _, ss := range in {
		if strings.EqualFold(ss, key) {
			return true
		}
	}
	return false
}

func InArray[T interface{}](val T, array []T) bool {
	s := reflect.ValueOf(array)
	for i := 0; i < s.Len(); i++ {
		if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
			return true
		}
	}
	return false
}

func IsArrayInArray[T interface{}](smallArray, largeArray []T) bool {
	for i := 0; i < len(smallArray); i++ {
		if !InArray[T](smallArray[i], largeArray) {
			return false
		}
	}
	return true
}

// DeepCopy 不支持导出结构体私有的变量，会报error
func DeepCopy[T interface{}](src T) (T, error) {
	// 创建一个缓冲区
	var buf bytes.Buffer
	var re T
	// 使用 Gob 编码将源数据编码到缓冲区
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(src)
	if err != nil {
		return re, err
	}

	// 创建一个与源数据类型相同的空白变量
	dest := reflect.New(reflect.TypeOf(src)).Interface()

	// 使用 Gob 解码将缓冲区的数据解码到新变量中
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(dest)
	if err != nil {
		return re, err
	}
	if v, ok := (dest).(*T); ok {
		re = *v
	} else if v, ok := (dest).(T); ok {
		re = v
	}
	return re, nil
}

// DeepCopyV2 无法导出私有字段会直接跳过
func DeepCopyV2(input interface{}) interface{} {
	if input == nil {
		return nil
	}
	in := reflect.Indirect(reflect.ValueOf(input))

	//// 如果输入是指针，则获取其指向的元素
	//if in.Kind() == reflect.Ptr {
	//	in = in.Elem()
	//}
	switch in.Kind() {

	case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		return input

	case reflect.Struct:
		out := reflect.Indirect(reflect.New(in.Type()))
		for i := 0; i < in.NumField(); i++ {
			if out.Field(i).IsValid() && out.Field(i).CanSet() {
				val := DeepCopyV2(in.Field(i).Interface())
				if val != nil {
					out.Field(i).Set(reflect.ValueOf(val))
				}
			}
		}
		return out.Interface()

	case reflect.Array, reflect.Slice:
		out := reflect.MakeSlice(in.Type(), in.Len(), in.Cap())
		for i := 0; i < in.Len(); i++ {
			if out.Field(i).CanSet() {
				val := DeepCopyV2(in.Index(i).Interface())
				if val != nil {
					out.Index(i).Set(reflect.ValueOf(val))
				}
			}
		}
		return out.Interface()

	case reflect.Map:
		out := reflect.MakeMapWithSize(in.Type(), in.Len())

		for _, key := range in.MapKeys() {
			out.SetMapIndex(DeepCopyV2(key.Interface()).(reflect.Value), DeepCopyV2(in.MapIndex(key).Interface()).(reflect.Value))
		}
		return out.Interface()
	default:
		return reflect.New(in.Type()).Elem().Interface()
	}
}

func B3TraceIdToUberTraceId(traceId, spanId, parentSpanID, flag string) string {
	if traceId == "" {
		return ""
	}
	if spanId == "" {
		spanId = "0000000000000000"
	}
	if parentSpanID == "" {
		parentSpanID = "0000000000000000"
	}
	if flag == "" {
		flag = "0"
	}
	return traceId + ":" + spanId + ":" + parentSpanID + ":" + flag
}

func SetContextWithB3(ctx context.Context, spanCtx jaeger.SpanContext) (reCtx context.Context) {
	if spanCtx.TraceID().String() == "" {
		return ctx
	}
	reCtx = InAndOutContextSetKV(ctx, "x-b3-traceid", spanCtx.TraceID().String())
	reCtx = InAndOutContextSetKV(reCtx, "x-b3-spanid", spanCtx.SpanID().String())
	reCtx = InAndOutContextSetKV(reCtx, "x-b3-sampled", fmt.Sprintf("%d", spanCtx.Flags()))
	reCtx = InAndOutContextSetKV(reCtx, "x-b3-parentspanid", spanCtx.ParentID().String())
	return
}

func UserHeaderToSetContextWithB3(ctx context.Context, header map[string][]string) (reCtx context.Context) {
	reCtx = ctx
	md := grpcMetadata.MD{}
	for key, vals := range header {
		md.Append(key, vals...)
	}
	for _, key := range []string{"x-b3-traceid", "x-b3-spanid", "x-b3-sampled", "x-b3-parentspanid"} {
		if val := md.Get(key); len(val) > 0 {
			reCtx = InAndOutContextSetKV(reCtx, key, val[0])
		}
	}
	return
}

func SetMdWithB3(md grpcMetadata.MD, openSpanCtx opentracing.SpanContext) grpcMetadata.MD {
	if md == nil {
		md = grpcMetadata.MD{}
	}
	if spanCtx, ok := openSpanCtx.(jaeger.SpanContext); ok {
		md.Set("x-b3-traceid", spanCtx.TraceID().String())
		md.Set("x-b3-spanid", spanCtx.SpanID().String())
		md.Set("x-b3-sampled", fmt.Sprintf("%d", spanCtx.Flags()))
		md.Set("x-b3-parentspanid", spanCtx.ParentID().String())
	}
	return md
}

func GetSpanCtxFromCtxB3KV(ctx context.Context) (opentracing.SpanContext, error) {
	// 先从服务调用的原始入口http/grpc（interceptor）拿b3信息继续链路写入/更新spanContext
	// 如果拿不到，则为链头开启新的trace并且写入B3信息到context
	md, _ := grpcMetadata.FromIncomingContext(ctx)
	var err error
	var spanCtx opentracing.SpanContext
	B3CombineTraceId, _ := GetB3CombineTraceIdAndUberTraceId(md)
	if B3CombineTraceId != "" {
		// 因为本项目只采用jaeger所以可以这么写，如果采用其他链路方案需要参考StartSpanWrapper的写法
		spanCtx, err = jaeger.ContextFromString(B3CombineTraceId)
		if err != nil {
			return spanCtx, err
		}
	}
	return spanCtx, nil
}

func GetSpanCtxFromCtx(ctx context.Context) (opentracing.SpanContext, error) {
	// 先从ctx中拿取spanContext
	// 如果拿不到，则为链头开启新的trace并且写入B3信息到context
	md, _ := grpcMetadata.FromIncomingContext(ctx)
	var err error
	var spanCtx opentracing.SpanContext
	B3CombineTraceId, _ := GetB3CombineTraceIdAndUberTraceId(md)
	if B3CombineTraceId != "" {
		// 因为本项目只采用jaeger所以可以这么写，如果采用其他链路方案需要参考StartSpanWrapper的写法
		spanCtx, err = jaeger.ContextFromString(B3CombineTraceId)
		if err != nil {
			return spanCtx, err
		}
	}
	return spanCtx, nil
}

func GetTraceIdFromCtx(ctx context.Context) string {
	if jaegerSpan, ok := opentracing.SpanFromContext(ctx).(*jaeger.Span); ok {
		return jaegerSpan.SpanContext().TraceID().String()
	}
	md, _ := grpcMetadata.FromIncomingContext(ctx)
	for key, values := range md {
		if len(values) == 0 {
			continue
		}
		if strings.EqualFold(key, "x-b3-traceid") {
			return values[0]
		} else if strings.EqualFold(key, "uber-trace-id") {
			uberTraceId := values[0]
			vecs := strings.Split(uberTraceId, ":")
			if len(vecs) > 0 {
				return vecs[0]
			}
		}
	}
	return ""
}

func ParseRPCErrorCode(err error) int64 {
	if s, ok := status.FromError(err); ok && s.Code() != codes.Unknown {
		return int64(s.Code())
	}
	result := gjson.Get(err.Error(), "code")
	if result.Exists() && result.Int() != 0 {
		return result.Int()
	}
	return int64(codes.Unknown)
}

func GetLessThenLimitStr(in []byte, limit int) string {
	strIn := string(in)
	if len(strIn) > limit {
		return strIn[:limit]
	}
	return strIn
}

func GetRequestKVParamStr(req *http.Request) string {
	queryParams := req.URL.Query()
	// 获取请求体中的参数
	formParams := req.PostForm

	// 将查询参数和请求体参数合并为一个 map
	//allParams := make(map[string]interface{})
	//for key, value := range queryParams {
	//	allParams[key] = value
	//}
	//for key, value := range formParams {
	//	allParams[key] = value
	//}
	allParams := lo.Assign(queryParams, formParams)
	if len(allParams) == 0 {
		return ""
	}
	// 将合并后的参数转换为 JSON 格式
	paramsJSON, err := json.Marshal(allParams)
	if err != nil {
		return ""
	}
	return string(paramsJSON)
}

func GetRequestBodyParamStr(req *http.Request) (re string) {
	content, err := io.ReadAll(req.Body)
	if err != nil {
		return ""
	}
	req.Body.Close()
	re = string(content)
	buffer := bytes.NewBuffer(content)
	req.Body = io.NopCloser(buffer)
	return
}

// 除去手机国际码后，手机号位数不少于10位时，只显示前三位和最后两位，如：129******66
// 手机号位数少于10位时，只显示前两位和后两位，如：15*****77
func MaskPhone(phone string) string {
	pre, suf := 0, 0
	if len(phone) >= 10 {
		pre, suf = 3, 2
	} else {
		pre, suf = 2, 2
	}
	rs := []rune(phone)
	for i := pre; i < len(rs)-suf; i++ {
		rs[i] = '*'
	}
	return string(rs)
}

func CamelToKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			// 如果是大写字母，且不是首字母，则在前面添加横杠
			if i > 0 {
				result = append(result, '-')
			}
			// 转换为小写字母
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func StringToMd5String(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

func ClearSyncMap(m *sync.Map) {
	m.Range(func(key, value interface{}) bool {
		m.Delete(key)
		return true
	})
}

func StorSyncMapFromNormalMap(normalMap map[string]string, m *sync.Map) {
	for key, val := range normalMap {
		m.Store(key, val)
	}
}

func GetNormalMapFromSyncMap(m *sync.Map) map[string]string {
	normalMap := map[string]string{}
	m.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if valStr, ok := value.(string); ok {
				normalMap[keyStr] = valStr
			}
		}
		return true
	})
	return normalMap
}

func GetSvcNameAndNsFromTarget(target string) (serviceName, namespace string, err error) {
	target = strings.TrimSuffix(target, ":50051")
	vec := strings.Split(target, ".")
	if len(vec) != 2 {
		err = errors.New("parse target failed")
		return
	}
	serviceName = vec[0]
	namespace = vec[1]
	return
}

func GetTarget(serviceName, namespace string) string {
	return serviceName + "." + namespace + ":50051"
}

func WriteStringToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func Ping(address string) bool {
	conn, err := net.DialTimeout("ip4:icmp", address, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func changeStu2Map(v interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		var name string
		tag := t.Field(i).Tag.Get("json")
		if tag == "-" {
			continue
		}
		if tag != "" {
			options := parseJSONTag(tag)
			name = options[0]
		} else {
			name = t.Field(i).Name
		}
		if rv.Field(i).Kind() == reflect.Struct || rv.Field(i).Kind() == reflect.Interface {
			data[name] = changeStu2Map(field.Interface())
		} else if rv.Field(i).Kind() == reflect.Slice {
			var slice []interface{}
			for j := 0; j < field.Len(); j++ {
				slice = append(slice, changeStu2Map(field.Index(j).Interface()))
			}
			data[name] = slice
		} else {
			data[name] = field.Interface()
		}
	}
	return data
}

func parseJSONTag(tag string) []string {
	return strings.Split(tag, ",")
}

func MarshWithOutOmitempty(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Interface && rv.Kind() != reflect.Slice {
		return json.Marshal(v)
	} else {
		return json.Marshal(changeStu2Map(v))
	}
}

func Sprintf(formatStr string, args ...any) string {
	if formatStr == "" {
		return ""
	}
	var result string
	reg := regexp.MustCompile(`\(\d+(%(\d*(\.)?\d*)?[a-z])\)`)
	placeholderList := reg.FindAllString(formatStr, -1)
	if len(placeholderList) == len(args) {
		list := make([]interface{}, 0, len(args))
		for _, placeholder := range placeholderList {
			start := strings.Index(placeholder, "%")
			placeholder = placeholder[1:start]
			index, err := strconv.Atoi(placeholder)
			if err != nil {
				return fmt.Sprintf(formatStr, args)
			}
			list = append(list, args[index-1])
		}
		ttm := reg.ReplaceAllString(formatStr, "${1}")
		result = fmt.Sprintf(ttm, list...)
	} else {
		result = fmt.Sprintf(formatStr, args...)
	}
	// 数字和中文之间增加空格/字母和数字之间增加空格
	reg = regexp.MustCompile("(\\d)([\u4e00-\u9fa5]|[a-zA-Z])")
	result = reg.ReplaceAllString(result, "${1} ${2}")
	reg = regexp.MustCompile("([\u4e00-\u9fa5]|[a-zA-Z])(\\d)")
	result = reg.ReplaceAllString(result, "${1} ${2}")
	// 去除多余的空格，多空格合并一个
	reg = regexp.MustCompile(`(\S)(\s+)(\S)`)
	result = reg.ReplaceAllString(result, "${1} ${3}")
	return strings.TrimSpace(result)
}
