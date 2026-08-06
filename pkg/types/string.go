package types

import (
	"encoding/json"
	"errors"
	"strconv"
)

type String string

func (s String) String() string        { return string(s) }
func (s String) Int64() (int64, error) { return strconv.ParseInt(string(s), 10, 64) }

func (s *String) UnmarshalJSON(src []byte) error {
	if string(src) == "null" || string(src) == `""` {
		return nil
	}

	var dst string
	if err := json.Unmarshal(src, &dst); err != nil {
		_, ok := errors.AsType[*json.UnmarshalTypeError](err)
		if !ok {
			return err
		}

		var num json.Number
		if err := json.Unmarshal(src, &num); err != nil {
			return err
		}
		dst = num.String()
	}

	*s = String(dst)

	return nil
}
