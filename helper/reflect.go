package helper

import "reflect"

// IsComparable returns whether v is not nil and has an underlying
// type that is comparable.
func IsComparable(v interface{}) bool {
	return v != nil && reflect.TypeOf(v).Comparable()
}

// IsSameType returns whether v1 and v2 are both not nil and have
// the same underlying type.
func IsSameType(v1 interface{}, v2 interface{}) bool {
	return v1 != nil && v2 != nil && reflect.TypeOf(v1) == reflect.TypeOf(v2)
}
