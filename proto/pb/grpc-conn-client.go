package pb

import (
	"context"
	"panda-go/common/grpc-wrapper/grpc_client_wrapper"
	"panda-go/common/log"
	"panda-go/common/utils"
	"reflect"
	"strings"
)

func main() {
	//NewGreeterClient()
}

func CreateClient[T interface{}](ctx context.Context) (interface{}, error) {
	var re any

	//types := reflect.TypeOf(main).PkgPath()
	//
	//// 遍历包中的所有类型
	//for i := 0; i < pkgValue.NumField(); i++ {
	//	field := pkgValue.Field(i)
	//	// 检查类型是否为结构体
	//	if field.Kind() == reflect.Struct {
	//		structName := field.Type().Name()
	//		fmt.Println("找到结构体:", structName)
	//	}
	//}

	//clientType := reflect.TypeOf((*T)(nil)).Elem()
	//client := reflect.New(clientType).Interface()
	//clientValue := reflect.ValueOf(client).Elem()
	//clientValue.MethodByName()
	//f := clientValue.FieldByName("cc")
	//if f.CanSet() {
	//	f.Set(reflect.ValueOf(conn-storage))
	//}
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t != nil {
		name := t.Name()
		name = utils.CamelToKebabCase(strings.TrimSuffix(name, "Client"))
		switch name {
		case "panda-account":
			conn, err := grpc_client_wrapper.GetConn(ctx, name)
			if err != nil {
				log.Error(ctx, "err")
				return nil, err
			}
			re = NewPandaAccountClient(conn)
		case "panda-auth":
			conn, err := grpc_client_wrapper.GetConn(ctx, name)
			if err != nil {
				log.Error(ctx, "err")
				return nil, err
			}
			re = NewPandaAuthClient(conn)
		case "panda-export":
			conn, err := grpc_client_wrapper.GetConn(ctx, name)
			if err != nil {
				log.Error(ctx, "err")
				return nil, err
			}
			re = NewCommonServiceClient(conn)
		}
	}
	return re, nil
}

func CreateClientWrapper[T interface{}](ctx context.Context) T {
	var x T
	tmp, _ := CreateClient[T](ctx)
	if cc, ok := tmp.(T); ok {
		x = cc
	}
	return x
}
