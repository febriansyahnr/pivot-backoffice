package httputil

import (
	"log"
	"net/http"
	"time"
)

// GetQueryParam retrieves the value of a query parameter from the URL of the given HTTP request.
// It returns the value of the query parameter and a boolean indicating whether the parameter was found.
func GetQueryParam(r *http.Request, key string) (string, bool) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return "", false
	}
	return v, true
}

// GetDateTimeQueryParam retrieves the value of a query parameter from the URL of the given HTTP request
// and parses it as a time.Time object. It returns the time.Time object and a boolean indicating whether
// the parameter was found and successfully parsed.
// when no explicit timezone in the param, then we assuming that the timezone come from UTC
func GetDateTimeQueryParam(r *http.Request, key string, layout string) (time.Time, bool) {
	var (
		t time.Time
	)

	v, ok := GetQueryParam(r, key)
	if !ok {
		return t, false
	}

	t, err := time.Parse(layout, v)
	if err != nil {
		log.Println("err-parse", err)
		return t, false
	}

	return t, true
}
