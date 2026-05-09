package utils

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

func NewUUID() string {
	u := uuid.New().String()
	return strings.Replace(u, "-", "", -1)
}

// mapToStruct 将 map 转换为结构体，根据 json tag 匹配字段
func mapToStruct[T any](m map[string]interface{}, tt T) (T, error) {
	var result T
	v := reflect.ValueOf(&result).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		// 获取 json tag 中的字段名（处理逗号分隔的情况）
		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		// 从 map 中获取值
		if val, ok := m[fieldName]; ok {
			fieldValue := v.Field(i)
			if fieldValue.CanSet() {
				valValue := reflect.ValueOf(val)
				if valValue.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(valValue)
				} else if valValue.Type().ConvertibleTo(fieldValue.Type()) {
					fieldValue.Set(valValue.Convert(fieldValue.Type()))
				}
			}
		}
	}

	return result, nil
}

// structToMap 将结构体转换为 map，使用 json tag 作为键名
func structToMap(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	elem := reflect.ValueOf(v)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	if elem.Kind() != reflect.Struct {
		return result
	}

	t := elem.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			} else if parts[0] == "-" {
				continue // 跳过 omitempty 等
			}
		}

		result[fieldName] = elem.Field(i).Interface()
	}

	return result
}

func ToMap[T []byte | any](data T) (map[string]interface{}, error) {
	switch v := any(data).(type) {
	case struct{}:
		return structToMap(v), nil
	case []byte:
		var m map[string]interface{}
		err := json.Unmarshal(v, &m)
		return m, err
	default:
		return nil, errors.New("not a map")
	}
}

func ToStruct[T map[string]any | []byte, R any](data T, r R) (R, error) {
	switch v := any(data).(type) {
	case map[string]interface{}:
		return mapToStruct(v, r)
	case []byte:
		err := json.Unmarshal(v, &r)
		return r, err
	default:
		var r R
		return r, errors.New("not a map")
	}
}

func ToJson[T map[string]any | any](data T) ([]byte, error) {
	return json.Marshal(data)
}
