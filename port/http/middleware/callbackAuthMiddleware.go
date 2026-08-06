package middleware

import (
	"errors"
	"net/http"
)

func MapRequiredHeader(r *http.Request, listHeader []string) (map[string]string, error) {
	headers := make(map[string]string)

	for _, key := range listHeader {

		data := r.Header.Get(key)
		if data == "" {
			// the error will return the missing header
			return headers, errors.New(key)
		}
		headers[key] = data
	}
	return headers, nil
}
