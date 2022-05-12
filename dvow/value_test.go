package dvow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverwriteValue_AsIs(t *testing.T) {
	scenarios := []struct {
		desc  string
		value interface{}
		want  interface{}
	}{
		{
			desc:  "string",
			value: "text",
			want:  "text",
		},
		{
			desc:  "bool",
			value: true,
			want:  true,
		},
		{
			desc:  "int",
			value: int(123),
			want:  int(123),
		},
		{
			desc:  "int8",
			value: int8(123),
			want:  int8(123),
		},
		{
			desc:  "int16",
			value: int16(123),
			want:  int16(123),
		},
		{
			desc:  "int32",
			value: int32(123),
			want:  int32(123),
		},
		{
			desc:  "int64",
			value: int64(123),
			want:  int64(123),
		},
		{
			desc:  "float32",
			value: float32(123.56),
			want:  float32(123.56),
		},
		{
			desc:  "float64",
			value: float64(123.45),
			want:  float64(123.45),
		},
		{
			desc:  "struct",
			value: struct{}{},
			want:  struct{}{},
		},
		{
			desc:  "slice",
			value: []struct{}{},
			want:  []struct{}{},
		},
		{
			desc:  "map",
			value: map[string]struct{}{},
			want:  map[string]struct{}{},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(sc.desc, func(t *testing.T) {
			sv := overwriteValue{
				value: sc.value,
			}

			actual := sv.AsIs()

			assert.Equal(t, sc.want, actual)
		})
	}
}

func TestOverwriteValue_AsString(t *testing.T) {
	scenarios := []struct {
		desc  string
		value interface{}
		want  string
	}{
		{
			desc:  "string",
			value: "text",
			want:  "text",
		},
		{
			desc:  "bool",
			value: true,
			want:  "",
		},
		{
			desc:  "int",
			value: int(123),
			want:  "",
		},
		{
			desc:  "int8",
			value: int8(123),
			want:  "",
		},
		{
			desc:  "int16",
			value: int16(123),
			want:  "",
		},
		{
			desc:  "int32",
			value: int32(123),
			want:  "",
		},
		{
			desc:  "int64",
			value: int64(123),
			want:  "",
		},
		{
			desc:  "float32",
			value: float32(123.56),
			want:  "",
		},
		{
			desc:  "float64",
			value: float64(123.45),
			want:  "",
		},
		{
			desc:  "struct",
			value: struct{}{},
			want:  "",
		},
		{
			desc:  "slice",
			value: []struct{}{},
			want:  "",
		},
		{
			desc:  "map",
			value: map[string]struct{}{},
			want:  "",
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(sc.desc, func(t *testing.T) {
			sv := overwriteValue{
				value: sc.value,
			}

			actual := sv.AsString()

			assert.Equal(t, sc.want, actual)
		})
	}
}

func TestOverwriteValue_AsBool(t *testing.T) {
	scenarios := []struct {
		desc  string
		value interface{}
		want  bool
	}{
		{
			desc:  "string",
			value: "text",
			want:  false,
		},
		{
			desc:  "bool",
			value: true,
			want:  true,
		},
		{
			desc:  "int",
			value: int(123),
			want:  false,
		},
		{
			desc:  "int8",
			value: int8(123),
			want:  false,
		},
		{
			desc:  "int16",
			value: int16(123),
			want:  false,
		},
		{
			desc:  "int32",
			value: int32(123),
			want:  false,
		},
		{
			desc:  "int64",
			value: int64(123),
			want:  false,
		},
		{
			desc:  "float32",
			value: float32(123.45),
			want:  false,
		},
		{
			desc:  "float64",
			value: float64(123.45),
			want:  false,
		},
		{
			desc:  "struct",
			value: struct{}{},
			want:  false,
		},
		{
			desc:  "slice",
			value: []struct{}{},
			want:  false,
		},
		{
			desc:  "map",
			value: map[string]struct{}{},
			want:  false,
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(sc.desc, func(t *testing.T) {
			sv := overwriteValue{
				value: sc.value,
			}

			actual := sv.AsBool()

			assert.Equal(t, sc.want, actual)
		})
	}
}

func TestOverwriteValue_AsFloat(t *testing.T) {
	scenarios := []struct {
		desc  string
		value interface{}
		want  float64
	}{
		{
			desc:  "string",
			value: "text",
			want:  0,
		},
		{
			desc:  "bool",
			value: true,
			want:  0,
		},
		{
			desc:  "int",
			value: int(123),
			want:  123,
		},
		{
			desc:  "int8",
			value: int8(123),
			want:  123,
		},
		{
			desc:  "int16",
			value: int16(123),
			want:  123,
		},
		{
			desc:  "int32",
			value: int32(123),
			want:  123,
		},
		{
			desc:  "int64",
			value: int64(123),
			want:  123,
		},
		{
			desc:  "float32",
			value: float32(123.45),
			want:  float64(float32(123.45)),
		},
		{
			desc:  "float64",
			value: float64(123.45),
			want:  123.45,
		},
		{
			desc:  "struct",
			value: struct{}{},
			want:  0,
		},
		{
			desc:  "slice",
			value: []struct{}{},
			want:  0,
		},
		{
			desc:  "map",
			value: map[string]struct{}{},
			want:  0,
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(sc.desc, func(t *testing.T) {
			sv := overwriteValue{
				value: sc.value,
			}

			actual := sv.AsFloat()

			assert.Equal(t, sc.want, actual)
		})
	}
}

func TestOverwriteValue_AsInt(t *testing.T) {
	scenarios := []struct {
		desc  string
		value interface{}
		want  int64
	}{
		{
			desc:  "string",
			value: "text",
			want:  0,
		},
		{
			desc:  "bool",
			value: true,
			want:  0,
		},
		{
			desc:  "int",
			value: int(123),
			want:  123,
		},
		{
			desc:  "int8",
			value: int8(123),
			want:  123,
		},
		{
			desc:  "int16",
			value: int16(123),
			want:  123,
		},
		{
			desc:  "int32",
			value: int32(123),
			want:  123,
		},
		{
			desc:  "int64",
			value: int64(123),
			want:  123,
		},
		{
			desc:  "float32",
			value: float32(123.56),
			want:  123,
		},
		{
			desc:  "float64",
			value: float64(123.45),
			want:  123,
		},
		{
			desc:  "struct",
			value: struct{}{},
			want:  0,
		},
		{
			desc:  "slice",
			value: []struct{}{},
			want:  0,
		},
		{
			desc:  "map",
			value: map[string]struct{}{},
			want:  0,
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(sc.desc, func(t *testing.T) {
			sv := overwriteValue{
				value: sc.value,
			}

			actual := sv.AsInt()

			assert.Equal(t, sc.want, actual)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(t *testing.T)
	}{
		{
			desc: "input containing field that cannot be marshalled",
			test: func(t *testing.T) {
				type dummy struct {
					Fn func()
				}

				sv := overwriteValue{
					value: dummy{},
				}

				result, err := Unmarshal[dummy](sv)

				assert.Nil(t, result)
				assert.Equal(t, "json: unsupported type: func()", err.Error())
			},
		},
		{
			desc: "valid input",
			test: func(t *testing.T) {
				type dummy struct {
					Text string
				}

				sv := overwriteValue{
					value: dummy{
						Text: "test",
					},
				}

				result, err := Unmarshal[dummy](sv)

				assert.Equal(t, &dummy{"test"}, result)
				assert.Nil(t, err)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(sc.desc, func(t *testing.T) {
			sc.test(t)
		})
	}
}
