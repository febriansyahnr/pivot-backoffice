package util

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueOfPointer(t *testing.T) {
	type payload struct {
		valString *string
		valInt    *int
		valFloat  *float64
		valBool   *bool
	}

	testCases := []struct {
		desc    string
		payload payload
		kind    types.BasicKind
		want    any
	}{
		{
			desc:    "when the data was nil, should return empty string",
			payload: payload{},
			kind:    types.String,
			want:    "",
		},
		{
			desc:    "when the data was nil, should return zero",
			payload: payload{},
			kind:    types.Int,
			want:    0,
		},
		{
			desc:    "when the data was nil, should return 0.0",
			payload: payload{},
			kind:    types.Float64,
			want:    0.0,
		},
		{
			desc:    "when the data was nil, should return false",
			payload: payload{},
			kind:    types.Bool,
			want:    false,
		},
		{
			desc:    "when the string pointer is not nil, should return the value",
			payload: payload{valString: newString("hello")},
			kind:    types.String,
			want:    "hello",
		},
		{
			desc:    "when the int pointer is not nil, should return the value",
			payload: payload{valInt: newInt(42)},
			kind:    types.Int,
			want:    42,
		},
		{
			desc:    "when the float pointer is not nil, should return the value",
			payload: payload{valFloat: newFloat64(3.14)},
			kind:    types.Float64,
			want:    3.14,
		},
		{
			desc:    "when the bool pointer is not nil, should return the value",
			payload: payload{valBool: newBool(true)},
			kind:    types.Bool,
			want:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			switch tc.kind {
			case types.String:
				val := ValueOfPtr(tc.payload.valString)
				assert.Equal(t, tc.want, val)
			case types.Int:
				val := ValueOfPtr(tc.payload.valInt)
				assert.Equal(t, tc.want, val)
			case types.Float64:
				val := ValueOfPtr(tc.payload.valFloat)
				assert.Equal(t, tc.want, val)
			case types.Bool:
				val := ValueOfPtr(tc.payload.valBool)
				assert.Equal(t, tc.want, val)
			}
		})
	}
}

func TestStructValueOfPtr(t *testing.T) {
	type child struct {
		Age int
	}

	type other struct {
		Name  string
		Child *child
	}

	t1 := other{}
	val := ValueOfPtr(t1.Child)
	assert.Equal(t, child{}, val)
}

func TestValueToPtr(t *testing.T) {
	type payload struct {
		valString string
		valInt    int
		valFloat  float64
		valBool   bool
	}

	testCases := []struct {
		desc    string
		payload payload
		kind    types.BasicKind
		want    any
	}{
		{
			desc:    "when the data was empty, should return pointer of empty string",
			payload: payload{},
			kind:    types.String,
			want:    "",
		},
		{
			desc:    "when the data was empty, should return zero",
			payload: payload{},
			kind:    types.Int,
			want:    0,
		},
		{
			desc:    "when the data was empty, should return 0.0",
			payload: payload{},
			kind:    types.Float64,
			want:    0.0,
		},
		{
			desc:    "when the data was empty, should return false",
			payload: payload{},
			kind:    types.Bool,
			want:    false,
		},
		{
			desc:    "when the string pointer is not nil, should return the value",
			payload: payload{valString: "hello"},
			kind:    types.String,
			want:    "hello",
		},
		{
			desc:    "when the int pointer is not nil, should return the value",
			payload: payload{valInt: 42},
			kind:    types.Int,
			want:    42,
		},
		{
			desc:    "when the float pointer is not nil, should return the value",
			payload: payload{valFloat: 3.14},
			kind:    types.Float64,
			want:    3.14,
		},
		{
			desc:    "when the bool pointer is not nil, should return the value",
			payload: payload{valBool: true},
			kind:    types.Bool,
			want:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			switch tc.kind {
			case types.String:
				val := ValueToPtr(tc.payload.valString)
				assert.NotNil(t, val)
				assert.Equal(t, tc.want, *val)
			case types.Int:
				val := ValueToPtr(tc.payload.valInt)
				assert.NotNil(t, val)
				assert.Equal(t, tc.want, *val)
			case types.Float64:
				val := ValueToPtr(tc.payload.valFloat)
				assert.NotNil(t, val)
				assert.Equal(t, tc.want, *val)
			case types.Bool:
				val := ValueToPtr(tc.payload.valBool)
				assert.NotNil(t, val)
				assert.Equal(t, tc.want, *val)
			}
		})
	}
}

// Helper functions to create pointers for test cases
func newString(val string) *string {
	return &val
}

func newInt(val int) *int {
	return &val
}

func newFloat64(val float64) *float64 {
	return &val
}

func newBool(val bool) *bool {
	return &val
}

func TestBoolPtr(t *testing.T) {
	p := BoolPtr(true) //nolint:modernize // testing deprecated BoolPtr

	assert.NotNil(t, p)
	assert.Equal(t, true, *p)

	p = BoolPtr(false) //nolint:modernize // testing deprecated BoolPtr
	assert.Equal(t, false, *p)
}

func TestClonePtr(t *testing.T) {
	// Test nil pointer
	result := ClonePtr[int](nil)
	assert.Nil(t, result)

	// Test integer
	n := 42
	p := &n
	clone := ClonePtr(p)

	assert.NotNil(t, clone)
	assert.NotSame(t, clone, p, "should return new pointer, not same as original")
	assert.Equal(t, 42, *clone)

	// Modify original, clone should be unaffected
	*p = 99
	assert.Equal(t, 42, *clone, "clone should be unaffected")

	// Test string
	s := "hello"
	pStr := &s
	cloneStr := ClonePtr(pStr)

	assert.Equal(t, "hello", *cloneStr)

	// Test struct
	type Person struct {
		Name string
		Age  int
	}
	person := Person{Name: "Alice", Age: 30}
	pPerson := &person
	clonePerson := ClonePtr(pPerson)

	assert.Equal(t, "Alice", clonePerson.Name)
	assert.Equal(t, 30, clonePerson.Age)

	// Modify original struct
	person.Name = "Bob"
	person.Age = 25
	assert.Equal(t, "Alice", clonePerson.Name, "clone.Name should be unaffected")
	assert.Equal(t, 30, clonePerson.Age, "clone.Age should be unaffected")
}
