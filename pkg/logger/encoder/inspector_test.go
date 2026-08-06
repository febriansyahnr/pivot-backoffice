package encoder_test

import (
	"encoding/json"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/logger/encoder"

	"github.com/stretchr/testify/assert"
)

func TestNewInspector(t *testing.T) {
	inspector := NewInspector([]string{"password"})

	data := map[string]interface{}{
		"id":       1,
		"username": "johnwick",
		"password": "abc123!@#",
	}
	want := `{"id": 1, "username":"johnwick", "password":"*********"}`

	buf, _ := json.Marshal(inspector.Inspects(data))

	if !assert.JSONEq(t, want, string(buf)) {
		t.Log("Output:", string(buf))
	}
}

func TestNewInspectorWithoutFields(t *testing.T) {
	inspector := NewInspector(nil)

	data := map[string]interface{}{
		"id":       1,
		"username": "johnwick",
		"password": "abc123!@#",
		"secret":   "ABC",
	}
	want := `{"id": 1, "username":"johnwick", "password":"abc123!@#", "secret":"ABC"}`

	buf, _ := json.Marshal(inspector.Inspects(data))

	if !assert.JSONEq(t, want, string(buf)) {
		t.Log("Output:", string(buf))
	}
}
