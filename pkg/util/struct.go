package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

func StructToString(data interface{}) (string, error) {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// Function to convert snake_case to camelCase
func SnakeToCamel(s string) string {
	var result string
	upper := false
	for i, v := range s {
		if v == '_' {
			upper = true
		} else {
			if upper {
				result += string(unicode.ToUpper(v))
				upper = false
			} else if i == 0 {
				result += string(unicode.ToLower(v))
			} else {
				result += string(v)
			}
		}
	}
	return result
}

// Function to convert camelCase to snake_case
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// MapSnakeToCamel Function to convert interface{} with possible map[string]interface{} or slice elements
func MapSnakeToCamel(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			// Convert the key
			newKey := SnakeToCamel(key)
			// Recursively process values
			newMap[newKey] = MapSnakeToCamel(value)
		}
		return newMap
	case []interface{}:
		// Handle slice of interfaces
		for i, elem := range v {
			v[i] = MapSnakeToCamel(elem)
		}
		return v
	default:
		rv := reflect.ValueOf(data)
		rt := reflect.TypeOf(data)

		switch rv.Kind() {
		case reflect.Struct:
			newMap := make(map[string]interface{})
			for i := 0; i < rv.NumField(); i++ {
				field := rt.Field(i)
				value := rv.Field(i)

				if !value.CanInterface() {
					continue
				}

				jsonTag := field.Tag.Get("json")
				fieldName := strings.Split(jsonTag, ",")[0]
				if fieldName == "" {
					fieldName = field.Name
				}

				newKey := SnakeToCamel(fieldName)
				newMap[newKey] = MapSnakeToCamel(value.Interface())
			}
			return newMap

		case reflect.Slice:
			var result []interface{}
			for i := 0; i < rv.Len(); i++ {
				elem := rv.Index(i)
				if elem.CanInterface() {
					result = append(result, MapSnakeToCamel(elem.Interface()))
				}
			}
			return result
		default:
			// If it's neither map nor slice, return the value as is
			return v
		}
	}
}

// MapCamelToSnake Function to convert interface{} with camelCase keys to snake_case keys
func MapCamelToSnake(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			// Convert the key
			newKey := CamelToSnake(key)
			// Recursively process values
			newMap[newKey] = MapCamelToSnake(value)
		}
		return newMap
	case []interface{}:
		// Handle slice of interfaces
		for i, elem := range v {
			v[i] = MapCamelToSnake(elem)
		}
		return v
	default:
		// If it's neither map nor slice, return the value as is
		return v
	}
}

// ConvertToStruct converts an interface{} to a specified struct type T using JSON marshaling and unmarshaling.
// This is useful for converting between similar struct types or when dealing with dynamic data.
// Example:
//
//	type User struct {
//	    Name string
//	    Age int
//	}
//
//	userMap := map[string]interface{}{"Name": "John", "Age": 30}
//	user, err := ConvertToStruct[User](userMap)
func ConvertToStruct[T any](data interface{}) (T, error) {
	var result T

	if data == nil {
		return result, fmt.Errorf("data is nil")
	}

	// check if data was string
	if str, ok := data.(string); ok {
		err := json.Unmarshal([]byte(str), &result)
		return result, err
	}

	// check if data was slice of byte
	if bytes, ok := data.([]byte); ok {
		err := json.Unmarshal(bytes, &result)
		return result, err
	}

	// Convert data to JSON and then unmarshal to struct
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return result, fmt.Errorf("error marshaling data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return result, fmt.Errorf("error unmarshaling data: %w", err)
	}

	return result, nil
}

func MergeStructToMap(m *map[string]any, s any) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	var temp map[string]any
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}
	for k, v := range temp {
		(*m)[k] = v
	}
	return nil
}
