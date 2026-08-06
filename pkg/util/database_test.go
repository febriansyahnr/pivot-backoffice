package util

import (
	"reflect"
	"testing"
	"time"
)

func TestScanJSON(t *testing.T) {
	type Person struct {
		Name     string    `json:"name"`
		Age      int       `json:"age"`
		Birthday time.Time `json:"birthday"`
	}

	tests := []struct {
		name    string
		value   interface{}
		want    Person
		wantErr bool
	}{
		{
			name:  "valid json",
			value: []byte(`{"name":"John Doe","age":30,"birthday":"2000-01-01T00:00:00Z"}`),
			want: Person{
				Name:     "John Doe",
				Age:      30,
				Birthday: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{"name":123}`),
			want:    Person{},
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   nil,
			want:    Person{},
			wantErr: false,
		},
		{
			name:    "wrong type",
			value:   "not a byte slice",
			want:    Person{},
			wantErr: true,
		},
		{
			name:    "empty json",
			value:   []byte(`{}`),
			want:    Person{},
			wantErr: false,
		},
		{
			name:    "malformed json",
			value:   []byte(`{"name":"John`),
			want:    Person{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dest Person
			err := ScanJSON(tt.value, &dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(dest, tt.want) {
				t.Errorf("ScanJSON() = %v, want %v", dest, tt.want)
			}
		})
	}
}
