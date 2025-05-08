package conf

import (
	"fmt"
	"reflect"
)

// Validator 配置验证器接口
type Validator interface {
	Validate() error
}

// ValidateConfig 验证配置
func ValidateConfig(config interface{}) error {
	// 检查是否实现了验证器接口
	if v, ok := config.(Validator); ok {
		return v.Validate()
	}

	// 基础验证
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 验证必填字段
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 检查 required 标签
		if fieldType.Tag.Get("required") == "true" {
			if field.IsZero() {
				return fmt.Errorf("field %s is required", fieldType.Name)
			}
		}
	}

	return nil
}
