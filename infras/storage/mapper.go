package storage

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func MapAttributes(entity interface{}) map[string]*string {
	if entity == nil {
		return nil
	}

	attributes := make(map[string]*string)
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Pointer {
		entityType = entityType.Elem()
	}

	value := reflect.ValueOf(entity)
	value = reflect.Indirect(value)
	for index := 0; index < entityType.NumField(); index++ {
		valueField := value.Field(index)
		attributeName := entityType.Field(index).Name
		if valueField.Kind() == reflect.Pointer {
			if valueField.IsNil() {
				attributes[attributeName] = nil
				continue
			}

			valueField = valueField.Elem()
		}

		attributeValue := valueField.Interface()
		attributeStr := toString(attributeValue)
		attributes[attributeName] = &attributeStr
	}

	return attributes
}

func toString(attributeValue interface{}) string {
	switch attributeValue.(type) {
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(attributeValue.(int64), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(attributeValue.(uint64), 10)
	case float32, float64:
		return strconv.FormatFloat(attributeValue.(float64), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(attributeValue.(bool))
	case time.Time:
		return attributeValue.(time.Time).Format(time.RFC3339)
	case string:
		return attributeValue.(string)
	default:
		return fmt.Sprintf("%v", attributeValue)
	}
}
