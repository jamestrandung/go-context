package dvow

import (
	"encoding/json"
)

//go:generate mockery --name Value --case underscore --inpkg
// Value wraps a raw interface{} value
type Value interface {
	// AsIs returns the wrapped value as-is.
	AsIs() interface{}
	// AsString typecast to string. Returns zero value if not possible to cast.
	AsString() string
	// AsBool typecast to bool. Returns zero value if not possible to cast.
	AsBool() bool
	// AsFloat typecast to float64. Returns zero value if not possible to cast.
	// Note: Try not to use a raw value of type float32 if possible.
	// https://stackoverflow.com/questions/67145364/golang-losing-precision-while-converting-float32-to-float64
	AsFloat() float64
	// AsInt typecast to int64. Returns zero value if not possible to cast.
	// NOTE: JSON by default unmarshal to numbers which are treated as float.
	// Using this method, your float will lose precision as an int64.
	AsInt() int64
}

type overwriteValue struct {
	value interface{}
}

// AsIs returns the wrapped value as-is.
func (v overwriteValue) AsIs() interface{} {
	return v.value
}

// AsString typecast to string. Returns zero value if not possible to cast.
func (v overwriteValue) AsString() (result string) {
	if castedValue, ok := (v.value).(string); ok {
		result = castedValue
	}

	return
}

// AsBool typecast to bool. Returns zero value if not possible to cast.
func (v overwriteValue) AsBool() (result bool) {
	if castedValue, ok := (v.value).(bool); ok {
		result = castedValue
	}

	return
}

// AsFloat typecast to float64. Returns zero value if not possible to cast.
// Note: Try not to use a raw value of type float32 if possible.
// https://stackoverflow.com/questions/67145364/golang-losing-precision-while-converting-float32-to-float64
func (v overwriteValue) AsFloat() (result float64) {
	switch v.value.(type) {
	case int:
		result = float64(v.value.(int))
	case int8:
		result = float64(v.value.(int8))
	case int16:
		result = float64(v.value.(int16))
	case int32:
		result = float64(v.value.(int32))
	case int64:
		result = float64(v.value.(int64))
	case float32:
		result = float64(v.value.(float32))
	case float64:
		result = v.value.(float64)
	}

	return
}

// AsInt typecast to int64. Returns zero value if not possible to cast.
// NOTE: JSON by default unmarshal to numbers which are treated as float.
// Using this method, your float will lose precision as an int64.
func (v overwriteValue) AsInt() (result int64) {
	switch v.value.(type) {
	case int:
		result = int64(v.value.(int))
	case int8:
		result = int64(v.value.(int8))
	case int16:
		result = int64(v.value.(int16))
	case int32:
		result = int64(v.value.(int32))
	case int64:
		result = v.value.(int64)
	case float32:
		result = int64(v.value.(float32))
	case float64:
		result = int64(v.value.(float64))
	}

	return
}

// Unmarshal into the given type
func Unmarshal[T any](v Value) (*T, error) {
	str, err := json.Marshal(v.AsIs())
	if err != nil {
		return nil, err
	}

	result := new(T)
	err = json.Unmarshal(str, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
