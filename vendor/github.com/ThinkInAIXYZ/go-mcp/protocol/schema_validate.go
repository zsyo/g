package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/ThinkInAIXYZ/go-mcp/pkg"
)

func VerifyAndUnmarshal(content json.RawMessage, v any) error {
	if len(content) == 0 {
		return fmt.Errorf("request arguments is empty")
	}

	t := reflect.TypeOf(v)
	for t.Kind() != reflect.Struct {
		if t.Kind() != reflect.Ptr {
			return fmt.Errorf("invalid type %v, plz use func `pkg.JSONUnmarshal` instead", t)
		}
		t = t.Elem()
	}

	typeUID := getTypeUUID(t)
	schema, ok := schemaCache.Load(typeUID)
	if !ok {
		return fmt.Errorf("schema has not been generated，unable to verify: plz use func `pkg.JSONUnmarshal` instead")
	}

	return verifySchemaAndUnmarshal(Property{
		Type:       ObjectT,
		Properties: schema.Properties,
		Required:   schema.Required,
	}, content, v)
}

func verifySchemaAndUnmarshal(schema Property, content []byte, v any) error {
	var data any
	err := pkg.JSONUnmarshal(content, &data)
	if err != nil {
		return err
	}
	if !validate(schema, data) {
		return errors.New("data validation failed against the provided schema")
	}
	return pkg.JSONUnmarshal(content, &v)
}

func validate(schema Property, data any) bool {
	switch schema.Type {
	case ObjectT:
		return validateObject(schema, data)
	case Array:
		return validateArray(schema, data)
	case String:
		str, ok := data.(string)
		if ok {
			return validateEnumProperty[string](str, schema.Enum, func(value string, enumValue string) bool {
				return value == enumValue
			})
		}
		return false
	case Number: // float64 and int
		if num, ok := data.(float64); ok {
			return validateEnumProperty[float64](num, schema.Enum, func(value float64, enumValue string) bool {
				if enumNum, err := strconv.ParseFloat(enumValue, 64); err == nil && value == enumNum {
					return true
				}
				return false
			})
		}
		if num, ok := data.(int); ok {
			return validateEnumProperty[int](num, schema.Enum, func(value int, enumValue string) bool {
				if enumNum, err := strconv.Atoi(enumValue); err == nil && value == enumNum {
					return true
				}
				return false
			})
		}
		return false
	case Boolean:
		_, ok := data.(bool)
		return ok
	case Integer:
		// Golang unmarshals all numbers as float64, so we need to check if the float64 is an integer
		if num, ok := data.(float64); ok {
			if num == float64(int64(num)) {
				return validateEnumProperty[float64](num, schema.Enum, func(value float64, enumValue string) bool {
					if enumNum, err := strconv.ParseFloat(enumValue, 64); err == nil && value == enumNum {
						return true
					}
					return false
				})
			}
			return false
		}

		if num, ok := data.(int); ok {
			return validateEnumProperty[int](num, schema.Enum, func(value int, enumValue string) bool {
				if enumNum, err := strconv.Atoi(enumValue); err == nil && value == enumNum {
					return true
				}
				return false
			})
		}

		if num, ok := data.(int64); ok {
			return validateEnumProperty[int64](num, schema.Enum, func(value int64, enumValue string) bool {
				if enumNum, err := strconv.Atoi(enumValue); err == nil && value == int64(enumNum) {
					return true
				}
				return false
			})
		}
		return false
	case Null:
		return data == nil
	default:
		return false
	}
}

func validateObject(schema Property, data any) bool {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return false
	}
	for _, field := range schema.Required {
		if _, exists := dataMap[field]; !exists {
			return false
		}
	}
	for key, valueSchema := range schema.Properties {
		value, exists := dataMap[key]
		if exists && !validate(*valueSchema, value) {
			return false
		}
	}
	return true
}

func validateArray(schema Property, data any) bool {
	dataArray, ok := data.([]any)
	if !ok {
		return false
	}
	for _, item := range dataArray {
		if !validate(*schema.Items, item) {
			return false
		}
	}
	return true
}

func validateEnumProperty[T any](data T, enum []string, compareFunc func(T, string) bool) bool {
	for _, enumValue := range enum {
		if compareFunc(data, enumValue) {
			return true
		}
	}
	return len(enum) == 0
}
