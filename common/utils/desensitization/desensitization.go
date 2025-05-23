package desensitization

import (
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"panda-go/common/utils"
	"reflect"
	"strings"
)

const MAX_TRIM_BYTES = 128

type dsRule struct {
	AffectFields []string
	Transform    func(string) string
}

var dsRuleMap = map[string]*dsRule{}

// implements dsRule ,and add more rule here
var dsRules = []*dsRule{
	{
		AffectFields: []string{"password", "Password"},
		Transform: func(raw string) string {
			return strings.Repeat("*", len(raw))
		},
	},
	{
		AffectFields: []string{"phone", "Phone"},
		Transform: func(raw string) string {
			return utils.MaskPhone(raw)
		},
	},
}

func init() {
	for _, r := range dsRules {
		for _, f := range r.AffectFields {
			dsRuleMap[f] = r
		}
	}
}

func FilterField(field *zap.Field) *zap.Field {
	switch field.Type {
	case zapcore.StringType:
		if r, ok := dsRuleMap[field.Key]; ok {
			f := zap.String(field.Key, r.Transform(field.String))
			return &f
		}
	case zapcore.ReflectType, zapcore.StringerType:
		if field.Interface != nil {
			v := desensitization(field.Interface)
			f := zap.Any(field.Key, v)
			return &f
		}
	}
	return field
}

func desensitization(val interface{}) interface{} {
	if val == nil || reflect.ValueOf(val).IsZero() {
		return val
	}
	var newVal interface{}
	if pm, ok := val.(proto.Message); ok {
		// proto.Clone is nearly 50 times faster than copier
		newVal = proto.Clone(pm)
	} else {
		valDirectType := reflect.Indirect(reflect.ValueOf(val)).Type()
		if valDirectType.Kind() != reflect.Struct {
			return val
		}
		//newVal = reflect.New(valDirectType).Interface()
		//_ = copier.CopyWithOption(newVal, val, copier.Option{DeepCopy: true})
		newVal = utils.DeepCopyV2(val)
	}

	findAndTransform(reflect.ValueOf(newVal))
	return newVal
}

func findAndTransform(val reflect.Value) {
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Struct:
		valType := val.Type()
		for i := 0; i < val.NumField(); i++ {
			f := reflect.Indirect(val.Field(i))
			if !f.CanSet() {
				continue
			}
			switch f.Kind() {
			case reflect.Struct:
				findAndTransform(f)
			case reflect.Slice:
				transformSlice(f)
			case reflect.String:
				str := f.String()
				rule := dsRuleMap[valType.Field(i).Name]
				if rule != nil {
					str = rule.Transform(str)
					f.Set(reflect.ValueOf(str))
				}
			}
		}
	case reflect.Slice:
		transformSlice(val)
	}
}

func transformSlice(value reflect.Value) {
	for j := 0; j < value.Len(); j++ {
		findAndTransform(value.Index(j))
	}
}
