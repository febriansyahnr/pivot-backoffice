package util_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestLookupMapAnyByKey(t *testing.T) {

	data := map[string]any{
		"key_s": "hello",
		"key_n": uint(1),
		"key_f": 2.5,
		"key_b": true,
	}

	valStr, ok := LookupMapAnyByKey[string](data, "key1")
	assert.False(t, ok)
	assert.Equal(t, "", valStr)

	valUInt, ok := LookupMapAnyByKey[uint](nil, "key_n")
	assert.False(t, ok)
	assert.Equal(t, uint(0), valUInt)

	valStr, ok = LookupMapAnyByKey[string](data, "key_s")
	assert.True(t, ok)
	assert.Equal(t, "hello", valStr)

	valUInt, ok = LookupMapAnyByKey[uint](&data, "key_n")
	assert.True(t, ok)
	assert.Equal(t, uint(1), valUInt)

	valFloat, ok := LookupMapAnyByKey[float64](&data, "key_f")
	assert.True(t, ok)
	assert.Equal(t, 2.5, valFloat)

	valBool, ok := LookupMapAnyByKey[bool](data, "key_b")
	assert.True(t, ok)
	assert.True(t, valBool)
}
