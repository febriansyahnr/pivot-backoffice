package util

import (
	"encoding/json"
	"errors"
	"fmt"
)

func JSONBind(dst, src any) (err error) {
	if dst == nil {
		return errors.New("destination pointer is nil - cannot bind source input")
	}

	var raw []byte

	switch val := src.(type) {
	case []byte:
		raw = val

	case string:
		raw = []byte(val)

	default:
		if raw, err = json.Marshal(src); err != nil {
			return fmt.Errorf("failed to marshal input data: %w", err)
		}
	}
	return json.Unmarshal(raw, dst)
}
