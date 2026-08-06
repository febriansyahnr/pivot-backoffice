package util

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeInt64ToUint64(t *testing.T) {
	tests := []struct {
		name    string
		input   int64
		want    uint64
		wantErr bool
	}{
		{
			name:    "positive value",
			input:   100,
			want:    100,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative value",
			input:   -100,
			want:    0,
			wantErr: true,
		},
		{
			name:    "max int64 value",
			input:   math.MaxInt64,
			want:    uint64(math.MaxInt64),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeInt64ToUint64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMustInt64ToUint64(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  uint64
	}{
		{
			name:  "positive value",
			input: 100,
			want:  100,
		},
		{
			name:  "zero value",
			input: 0,
			want:  0,
		},
		{
			name:  "negative value",
			input: -100,
			want:  0, // Should return 0 for negative values
		},
		{
			name:  "max int64 value",
			input: math.MaxInt64,
			want:  uint64(math.MaxInt64),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustInt64ToUint64(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSafeUint64ToInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   uint64
		want    int64
		wantErr bool
	}{
		{
			name:    "small value",
			input:   100,
			want:    100,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "max int64 value",
			input:   uint64(math.MaxInt64),
			want:    math.MaxInt64,
			wantErr: false,
		},
		{
			name:    "exceeds max int64",
			input:   uint64(math.MaxInt64) + 1,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeUint64ToInt64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMustUint64ToInt64(t *testing.T) {
	tests := []struct {
		name  string
		input uint64
		want  interface{}
	}{
		{
			name:  "small value",
			input: 100,
			want:  int64(100),
		},
		{
			name:  "zero value",
			input: 0,
			want:  int64(0),
		},
		{
			name:  "max int64 value",
			input: uint64(math.MaxInt64),
			want:  int64(math.MaxInt64),
		},
		{
			name:  "exceeds max int64",
			input: uint64(math.MaxInt64) + 1,
			want:  "9223372036854775808",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustUint64ToInt64(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSafeInt32ToUint32(t *testing.T) {
	tests := []struct {
		name    string
		input   int32
		want    uint32
		wantErr bool
	}{
		{
			name:    "positive value",
			input:   100,
			want:    100,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative value",
			input:   -100,
			want:    0,
			wantErr: true,
		},
		{
			name:    "max int32 value",
			input:   math.MaxInt32,
			want:    uint32(math.MaxInt32),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeInt32ToUint32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSafeUint32ToInt32(t *testing.T) {
	tests := []struct {
		name    string
		input   uint32
		want    int32
		wantErr bool
	}{
		{
			name:    "small value",
			input:   100,
			want:    100,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "max int32 value",
			input:   uint32(math.MaxInt32),
			want:    math.MaxInt32,
			wantErr: false,
		},
		{
			name:    "exceeds max int32",
			input:   uint32(math.MaxInt32) + 1,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeUint32ToInt32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSafeIntToUint16(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    uint16
		wantErr bool
	}{
		{
			name:    "positive value",
			input:   100,
			want:    100,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative value",
			input:   -100,
			want:    0,
			wantErr: true,
		},
		{
			name:    "max uint16 value",
			input:   math.MaxUint16,
			want:    math.MaxUint16,
			wantErr: false,
		},
		{
			name:    "exceeds max uint16 value",
			input:   math.MaxUint16 + 1,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeIntToUint16(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMustIntToUint16(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  uint16
	}{
		{
			name:  "positive value",
			input: 100,
			want:  100,
		},
		{
			name:  "zero value",
			input: 0,
			want:  0,
		},
		{
			name:  "negative value",
			input: -100,
			want:  0, // Should return 0 for negative values
		},
		{
			name:  "max uint16 value",
			input: math.MaxUint16,
			want:  math.MaxUint16,
		},
		{
			name:  "exceeds max uint16 value",
			input: math.MaxUint16 + 1,
			want:  math.MaxUint16, // Should clamp to max value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustIntToUint16(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSafeIntToUint32(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    uint32
		wantErr bool
	}{
		{
			name:    "positive value",
			input:   100,
			want:    100,
			wantErr: false,
		},
		{
			name:    "zero value",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative value",
			input:   -100,
			want:    0,
			wantErr: true,
		},
		{
			name:    "max int32 value",
			input:   math.MaxInt32,
			want:    uint32(math.MaxInt32),
			wantErr: false,
		},
		{
			name:    "max uint32 value",
			input:   math.MaxUint32,
			want:    math.MaxUint32,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeIntToUint32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMustIntToUint32(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  uint32
	}{
		{
			name:  "positive value",
			input: 100,
			want:  100,
		},
		{
			name:  "zero value",
			input: 0,
			want:  0,
		},
		{
			name:  "negative value",
			input: -100,
			want:  0, // Should return 0 for negative values
		},
		{
			name:  "max int32 value",
			input: math.MaxInt32,
			want:  uint32(math.MaxInt32),
		},
		{
			name:  "max uint32 value",
			input: math.MaxUint32,
			want:  math.MaxUint32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MustIntToUint32(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
