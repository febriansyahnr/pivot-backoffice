package util_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timeLocal, _ = time.LoadLocation("Asia/Jakarta")

func mockLocationLoader(name string) (*time.Location, error) {
	return nil, errors.New("mock location error")
}

func TestGetTimeWithLoader(t *testing.T) {
	tests := []struct {
		name       string
		loaderFunc util.LocationLoader
		wantErr    bool
	}{
		{
			name:       "Successful Timezone Retrieval",
			loaderFunc: time.LoadLocation,
			wantErr:    false,
		},
		{
			name:       "Error Handling",
			loaderFunc: mockLocationLoader,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := util.GetJakartaTimeWithLoader(tt.loaderFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: GetJakartaTimeWithLoader() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestGetJakartaTime(t *testing.T) {
	jakartaTime, err := util.GetJakartaTime()
	if err != nil {
		t.Errorf("GetJakartaTime() returned an error: %v", err)
	}

	_, offset := jakartaTime.Zone()
	if offset != 7*60*60 {
		t.Errorf("GetJakartaTime() did not return time in 'Asia/Jakarta' timezone: got offset %d", offset)
	}
}

func TestConstSnapDateFormatLayout(t *testing.T) {
	dateStr := "2024-05-08T12:34:05+07:00"

	date, err := time.ParseInLocation(util.SnapDateFormatLayout, dateStr, timeLocal)

	require.Nil(t, err)
	require.Equal(t, 2024, date.Year())
	require.Equal(t, time.Month(5), date.Month())
	require.Equal(t, 8, date.Day())
	require.Equal(t, 12, date.Hour())
	require.Equal(t, 34, date.Minute())
	require.Equal(t, 5, date.Second())
}

func TestConvertToJakarta(t *testing.T) {
	utc := time.Date(2024, 5, 8, 8, 32, 12, 0, time.UTC)
	expected := time.Date(2024, 5, 8, 15, 32, 12, 0, timeLocal)

	assert.Equal(t, expected, util.ConvertToJakarta(utc))
}

func TestSnapCompatible(t *testing.T) {
	tests := []struct {
		utc  time.Time
		want string
	}{
		{
			utc:  time.Date(2024, 5, 8, 8, 32, 12, 0, time.UTC),
			want: "2024-05-08T15:32:12+07:00",
		},
		{
			utc:  time.Date(2024, 5, 4, 6, 25, 55, 0, time.UTC),
			want: "2024-05-04T13:25:55+07:00",
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, util.SnapCompatible(test.utc))
	}
}

func TestGetCurrentTimeWithMillisFormatted(t *testing.T) {
	nowStr := time.Now().Format("20060102")
	assert.Equal(t, nowStr, util.GetCurrentTimeWithMillisFormatted()[:8])
}

func TestDateMonthStrYear(t *testing.T) {
	tests := []struct {
		input time.Time
		want  string
	}{
		{
			input: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			want:  "31 January 2024",
		},
		{
			input: time.Date(2024, 6, 24, 0, 0, 0, 0, time.UTC),
			want:  "24 June 2024",
		},
		{
			input: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			want:  "01 June 2023",
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, util.DateStrMonthYear(test.input))
	}
}

func TestMonthName(t *testing.T) {
	tests := map[time.Time]string{
		time.Date(2024, 6, 24, 0, 0, 0, 0, time.UTC): "June",
		time.Date(2024, 8, 31, 0, 0, 0, 0, time.UTC): "August",
		time.Date(2024, 10, 3, 0, 0, 0, 0, time.UTC): "October",
		time.Date(2024, 3, 3, 0, 0, 0, 0, time.UTC):  "March",
	}
	for input, want := range tests {
		assert.Equal(t, want, util.MonthName(input))
	}
}

func TestDateStrMonthYearHour(t *testing.T) {
	tests := []struct {
		input time.Time
		want  string
	}{
		{
			input: time.Date(2024, 7, 29, 14, 48, 33, 0, time.UTC),
			want:  "29 July 2024 09:48:33 PM",
		},
		{
			input: time.Date(2024, 2, 15, 8, 55, 12, 0, time.UTC),
			want:  "15 February 2024 03:55:12 PM",
		},
		{
			input: time.Date(2023, 12, 14, 20, 57, 24, 0, time.UTC),
			want:  "15 December 2023 03:57:24 AM",
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, util.DateStrMonthYearHour(test.input))
	}
}
func TestGetCurrentDateOfLocation(t *testing.T) {
	tests := []struct {
		name     string
		location *time.Location
	}{
		{
			name:     "UTC Location",
			location: time.UTC,
		},
		{
			name:     "Jakarta Location",
			location: timeLocal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.GetCurrentDateOfLocation(tt.location)
			now := time.Now().In(tt.location)

			assert.Equal(t, now.Year(), result.Year())
			assert.Equal(t, now.Month(), result.Month())
			assert.Equal(t, now.Day(), result.Day())
			assert.Equal(t, 0, result.Hour())
			assert.Equal(t, 0, result.Minute())
			assert.Equal(t, 0, result.Second())
			assert.Equal(t, 0, result.Nanosecond())
			assert.Equal(t, tt.location, result.Location())
		})
	}
}
func TestGetJakartaTimeLocation(t *testing.T) {
	loc := util.GetJakartaTimeLocation()
	assert.NotNil(t, loc)
	assert.Equal(t, "Asia/Jakarta", loc.String())
}

func TestConvertToJakartaString(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "UTC time converted to Jakarta time",
			input:    time.Date(2024, 11, 21, 7, 35, 26, 0, time.UTC),
			expected: "21/11/2024 14:35:26 WIB",
		},
		{
			name:     "Midnight UTC converted to Jakarta time",
			input:    time.Date(2024, 5, 8, 0, 0, 0, 0, time.UTC),
			expected: "08/05/2024 07:00:00 WIB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.ConvertToJakartaString(tt.input)
			assert.Equal(t, tt.expected, result, "Expected formatted Jakarta time does not match")
		})
	}
}

func TestSnapTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "SUCCESS: valid time conversion",
			input:    time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
			expected: "2023-05-15T21:30:00+07:00",
		},
		{
			name:     "SUCCESS: midnight time conversion",
			input:    time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			expected: "2023-05-15T07:00:00+07:00",
		},
		{
			name:     "SUCCESS: leap year date conversion",
			input:    time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			expected: "2024-02-29T19:00:00+07:00",
		},
		{
			name:     "SUCCESS: year boundary conversion",
			input:    time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2024-01-01T06:59:59+07:00",
		},
		{
			name:     "SUCCESS: with nanoseconds",
			input:    time.Date(2023, 5, 15, 14, 30, 0, 500000000, time.UTC),
			expected: "2023-05-15T21:30:00+07:00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := util.SnapTimeFormat(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeToUTC(t *testing.T) {
	tests := []struct {
		name      string
		inputTime time.Time
		timezone  string
		expected  time.Time
		wantErr   bool
	}{
		{
			name:      "Valid Jakarta timezone",
			inputTime: time.Date(2024, 6, 1, 15, 30, 0, 0, timeLocal),
			timezone:  "Asia/Jakarta",
			expected:  time.Date(2024, 6, 1, 8, 30, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "Valid New York timezone",
			inputTime: time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("EDT", -4*60*60)),
			timezone:  "America/New_York",
			expected:  time.Date(2024, 6, 1, 16, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "Valid London timezone",
			inputTime: time.Date(2024, 6, 1, 13, 45, 30, 0, time.FixedZone("BST", 1*60*60)),
			timezone:  "Europe/London",
			expected:  time.Date(2024, 6, 1, 12, 45, 30, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "Invalid timezone",
			inputTime: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
			timezone:  "Invalid/Timezone",
			expected:  time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
			wantErr:   true,
		},
		{
			name:      "UTC timezone",
			inputTime: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
			timezone:  "UTC",
			expected:  time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "Zero time",
			inputTime: time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			timezone:  "UTC",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := util.TimeToUTC(tt.inputTime, tt.timezone)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
func TestGetTimeZoneFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    *time.Location
		wantErr bool
	}{
		{
			name:    "Nil context",
			ctx:     nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Context without timezone",
			ctx:     context.Background(),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Context with invalid timezone",
			ctx:     context.WithValue(context.Background(), constant.CtxTimeZone, "Invalid/Timezone"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Context with valid timezone (Asia/Jakarta)",
			ctx:     context.WithValue(context.Background(), constant.CtxTimeZone, "Asia/Jakarta"),
			want:    timeLocal,
			wantErr: false,
		},
		{
			name:    "Context with valid timezone (UTC)",
			ctx:     context.WithValue(context.Background(), constant.CtxTimeZone, "UTC"),
			want:    time.UTC,
			wantErr: false,
		},
		{
			name:    "Context with valid timezone (Europe/London)",
			ctx:     context.WithValue(context.Background(), constant.CtxTimeZone, "Europe/London"),
			want:    func() *time.Location { loc, _ := time.LoadLocation("Europe/London"); return loc }(),
			wantErr: false,
		},
		{
			name:    "Context with non-string timezone value",
			ctx:     context.WithValue(context.Background(), constant.CtxTimeZone, 123),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.GetTimeZoneFromContext(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.String(), got.String())
			}
		})
	}
}

func TestParseTimeToDatetime(t *testing.T) {
	loc, err := time.LoadLocation(constant.TimeLoc)
	require.NoError(t, err)

	date := time.Date(2026, 2, 10, 11, 0, 0, 0, loc)

	tests := []struct {
		time         string
		wantDatetime time.Time
	}{
		{
			time:         "07:00:00",
			wantDatetime: time.Date(2026, 2, 10, 7, 0, 0, 0, loc),
		},
		{
			time:         "12:30:59",
			wantDatetime: time.Date(2026, 2, 10, 12, 30, 59, 0, loc),
		},
		{
			time:         "17:01:30",
			wantDatetime: time.Date(2026, 2, 10, 17, 1, 30, 0, loc),
		},
		{
			time:         "19:15:40",
			wantDatetime: time.Date(2026, 2, 10, 19, 15, 40, 0, loc),
		},
		{
			time:         "23:29:11",
			wantDatetime: time.Date(2026, 2, 10, 23, 29, 11, 0, loc),
		},
	}
	for _, test := range tests {
		datetime, err := util.ParseTimeToDatetime(date, test.time)
		require.NoError(t, err)
		require.Equal(t, test.wantDatetime, datetime)
	}
}
