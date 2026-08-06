package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

const (
	maxBackdateMonths = 6
	maxDateRangeDays  = 31
)

type ErrInvalidFormat struct{ message string }

func (e ErrInvalidFormat) Error() string {
	return e.message
}

type ErrValidation struct{ message string }

func (e ErrValidation) Error() string {
	return e.message
}

type Getter interface {
	Get(key string) string
}

type KeyValues struct {
	data map[string]any
}

func (k *KeyValues) Get(key string) string {
	val, _ := k.data[key].(string)
	return val
}

func (k *KeyValues) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	return json.Unmarshal(data, &k.data)
}

// ValidateReportDateRangeFromRequest validates a reporting date range taken from an HTTP request.
//
// Purpose
//   - Validate a start/end date range for reporting endpoints (merchant dashboard, reports, etc).
//   - Support extracting the date range either from query parameters or from JSON request body.
//   - Enforce two business constraints:
//     1) maximum allowed range length (maxRangeDays) — e.g. 31 days
//     2) maximum backdate window (maxBackdateMonths) — e.g. 6 months
//
// Example
//
//	ValidateReportDateRangeFromRequest(r, "startDate", "endDate", "paymentStartDate", "paymentEndDate")
func ValidateReportDateRangeFromRequest(request any, keyPairs ...string) error {
	if (len(keyPairs) % 2) != 0 {
		return errors.New("parameters must consist of paired initial and final values")
	}

	pairs := slices.Chunk(keyPairs, 2)

	times := map[int]string{
		0: "T00:00:00Z", 1: "T23:59:59Z",
	}
	parse := func(i int, d string) (time.Time, error) {
		if len(d) == 10 {
			d = d + times[i]
		}
		return time.Parse(time.RFC3339, d) // Parse string datetime with including timezone
	}

	now := time.Now().UTC()
	maxBackdate := now.AddDate(0, -maxBackdateMonths, 0)

	var params Getter

	switch r := request.(type) {
	case *http.Request:
		params = r.URL.Query()

	default:
		raw, _ := json.Marshal(r)

		params = &KeyValues{}
		if err := json.Unmarshal(raw, params); err != nil {
			return fmt.Errorf("json unmarshal request: %v", err)
		}
	}

	for pair := range pairs {

		k0, k1 := pair[0], pair[1]

		s0, s1 := params.Get(k0), params.Get(k1)
		if s0 == "" || s1 == "" || strings.Contains(s0, "0001-01-01") || strings.Contains(s1, "0001-01-01") {
			continue
		}

		t0, err := parse(0, s0)
		if err != nil {
			return &ErrInvalidFormat{
				message: fmt.Sprintf("Key: %s Value: %s Error: Value format must be yyyy-mm-ddThh:nn:ssZ", k0, s0),
			}
		}

		t1, err := parse(1, s1)
		if err != nil {
			return &ErrInvalidFormat{
				message: fmt.Sprintf("Key: %s Value: %s Error: Value format must be yyyy-mm-ddThh:nn:ssZ", k1, s1),
			}
		}

		if t0.After(t1) {
			return &ErrValidation{message: fmt.Sprintf("%s must not be greater than %s", k0, k1)}

		} else if t0.Before(maxBackdate) {
			return &ErrValidation{message: fmt.Sprintf("The date range exceeds the allowed backdate limit. Maximum allowed is the last %d months.", maxBackdateMonths)}

		} else if (t1.Sub(t0).Hours() / 24) > maxDateRangeDays {
			return &ErrValidation{message: fmt.Sprintf("The date range exceeds the allowed limit. Maximum permitted is %d days.", maxDateRangeDays)}
		}
	}
	return nil
}
