package config

import "strings"

type MStrStr map[string]string

func (m *MStrStr) KeyToUpperCase() {
	if m == nil {
		return
	}
	dst := MStrStr{}
	for key, val := range *m {
		dst[strings.ToUpper(key)] = val
	}
	*m = dst
}
