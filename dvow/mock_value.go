// Code generated by mockery v1.0.0. DO NOT EDIT.

package dvow

import mock "github.com/stretchr/testify/mock"

// MockValue is an autogenerated mock type for the Value type
type MockValue struct {
	mock.Mock
}

// AsBool provides a mock function with given fields:
func (_m *MockValue) AsBool() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// AsFloat provides a mock function with given fields:
func (_m *MockValue) AsFloat() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// AsInt provides a mock function with given fields:
func (_m *MockValue) AsInt() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// AsIs provides a mock function with given fields:
func (_m *MockValue) AsIs() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// AsString provides a mock function with given fields:
func (_m *MockValue) AsString() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Unmarshal provides a mock function with given fields: t
func (_m *MockValue) Unmarshal(t interface{}) error {
	ret := _m.Called(t)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(t)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
