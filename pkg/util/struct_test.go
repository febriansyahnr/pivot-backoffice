package util

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func TestStructToString(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Field1: "value1",
				Field2: 42,
			},
			want: `{
  "field1": "value1",
  "field2": 42
}`,
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    "null",
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   func() {},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructToString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StructToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestConvertToStruct(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	tests := []struct {
		name    string
		input   interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:    "map to struct",
			input:   map[string]interface{}{"name": "Alice", "age": 30},
			want:    Person{Name: "Alice", Age: 30},
			wantErr: false,
		},
		{
			name:    "nested struct conversion",
			input:   map[string]interface{}{"street": "Main St", "city": "New York", "country": "USA"},
			want:    Address{Street: "Main St", City: "New York", Country: "USA"},
			wantErr: false,
		},
		{
			name:    "incompatible types",
			input:   []int{1, 2, 3},
			want:    Person{},
			wantErr: true,
		},
		{
			name:    "missing fields",
			input:   map[string]interface{}{"name": "Dave"},
			want:    Person{Name: "Dave", Age: 0},
			wantErr: false,
		},
		{
			name:    "valid JSON string",
			input:   `{"field1":"test value","field2":123}`,
			want:    TestStruct{Field1: "test value", Field2: 123},
			wantErr: false,
		},
		{
			name:    "invalid JSON string",
			input:   `{"field1":"broken"`,
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name:    "empty JSON object",
			input:   `{}`,
			want:    TestStruct{},
			wantErr: false,
		},
		{
			name:    "valid JSON bytes",
			input:   []byte(`{"field1":"byte data","field2":456}`),
			want:    TestStruct{Field1: "byte data", Field2: 456},
			wantErr: false,
		},
		{
			name:    "invalid JSON bytes",
			input:   []byte(`not a json`),
			want:    TestStruct{},
			wantErr: true,
		},
		{
			name:    "when the payload was nil, then retur error",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case Person:
				got, err := ConvertToStruct[Person](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConvertToStruct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want.(Person) {
					t.Errorf("ConvertToStruct() = %v, want %v", got, tt.want)
				}
			case map[string]interface{}:
				got, err := ConvertToStruct[map[string]interface{}](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConvertToStruct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					// Compare keys and values
					expected := tt.want.(map[string]interface{})
					if len(got) != len(expected) {
						t.Errorf("ConvertToStruct() = %v, want %v", got, expected)
					}
					for k, v := range expected {
						if got[k] != v {
							t.Errorf("ConvertToStruct() = %v, want %v", got, expected)
						}
					}
				}
			case Address:
				got, err := ConvertToStruct[Address](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConvertToStruct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want.(Address) {
					t.Errorf("ConvertToStruct() = %v, want %v", got, tt.want)
				}
			case TestStruct:
				got, err := ConvertToStruct[TestStruct](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConvertToStruct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ConvertToStruct() = %v, want %v", got, tt.want)
				}
			case []byte:
				got, err := ConvertToStruct[TestStruct](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConvertToStruct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ConvertToStruct() = %v, want %v", got, tt.want)
				}
			default:
				got, err := ConvertToStruct[interface{}](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ConvertToStruct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ConvertToStruct() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
func TestMapSnakeToCamel(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "empty map",
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
		{
			name:  "simple map",
			input: map[string]interface{}{"first_name": "John", "last_name": "Doe"},
			want:  map[string]interface{}{"firstName": "John", "lastName": "Doe"},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"user_info": map[string]interface{}{
					"first_name": "John",
					"last_name":  "Doe",
				},
			},
			want: map[string]interface{}{
				"userInfo": map[string]interface{}{
					"firstName": "John",
					"lastName":  "Doe",
				},
			},
		},
		{
			name: "map with array",
			input: map[string]interface{}{
				"user_roles": []interface{}{
					map[string]interface{}{"role_name": "admin", "role_id": 1},
					map[string]interface{}{"role_name": "user", "role_id": 2},
				},
			},
			want: map[string]interface{}{
				"userRoles": []interface{}{
					map[string]interface{}{"roleName": "admin", "roleId": 1},
					map[string]interface{}{"roleName": "user", "roleId": 2},
				},
			},
		},
		{
			name:  "primitive value",
			input: "hello_world",
			want:  "hello_world", // Primitive values remain unchanged
		},
		{
			name:  "slice of primitives",
			input: []interface{}{1, "two", 3.0},
			want:  []interface{}{1, "two", 3.0}, // Primitive values in slice remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapSnakeToCamel(tt.input)

			// Use deep equality check for complex structures
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapSnakeToCamel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple snake case",
			input: "hello_world",
			want:  "helloWorld",
		},
		{
			name:  "multiple underscores",
			input: "user_first_name",
			want:  "userFirstName",
		},
		{
			name:  "ending with underscore",
			input: "trailing_underscore_",
			want:  "trailingUnderscore",
		},
		{
			name:  "consecutive underscores",
			input: "multiple__underscores",
			want:  "multipleUnderscores",
		},
		{
			name:  "single character segments",
			input: "a_b_c",
			want:  "aBC",
		},
		{
			name:  "no underscores",
			input: "plain",
			want:  "plain",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnakeToCamel(tt.input); got != tt.want {
				t.Errorf("SnakeToCamel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeStructToMap(t *testing.T) {
	existing := map[string]any{
		"existing_key": "existing_value",
	}

	structTest := struct {
		Test    string `json:"test"`
		Message string `json:"message"`
	}{
		Test:    "test",
		Message: "test message",
	}

	err := MergeStructToMap(&existing, structTest)
	require.NoError(t, err)

	require.Equal(t, "existing_value", existing["existing_key"])
	require.Equal(t, "test", existing["test"])
	require.Equal(t, "test message", existing["message"])
}
