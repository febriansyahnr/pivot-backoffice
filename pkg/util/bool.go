package util

import (
	"net/http"
	"strconv"
	"strings"
)

func PostFormBoolValue(r *http.Request, key string) (bool, error) {
	val := strings.TrimSpace(
		strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue(key), `"`, ``), `'`, ``),
	)
	if val == "" {
		val = "false"
	}
	return strconv.ParseBool(val)
}
