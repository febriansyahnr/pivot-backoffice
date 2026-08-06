package util

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

const (
	UTCLayout            = "2006-01-02T15:04:05Z"
	SnapDateFormatLayout = "2006-01-02T15:04:05+07:00"
	ISO8601FormatLayout  = "2006-01-02 15:04:05"
	ReceiptFormatLayout  = "02 Jan 2006, 15:04:05"
)

var (
	TimeNow = time.Now().UTC() // please use only for unit test needs
)

var loc, _ = time.LoadLocation(constant.TimeLoc)

type LocationLoader func(name string) (*time.Location, error)

func GetJakartaTimeWithLoader(loader LocationLoader) (time.Time, error) {
	t, err := loader("Asia/Jakarta")
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().In(t), nil
}

func GetJakartaTime() (time.Time, error) {
	return GetJakartaTimeWithLoader(time.LoadLocation)
}

func ConvertToJakarta(oldTime time.Time) time.Time {
	t, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Println("error when load location", err)
		return time.Time{}
	}

	return oldTime.In(t)
}

func SnapCompatible(oldTime time.Time) string {
	t, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Println("error when load location", err)
		return ""
	}
	return oldTime.In(t).Format(SnapDateFormatLayout)
}

func SnapTimeFormat(oldTime time.Time) string {
	t, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Println("error when load location", err)
		return ""
	}

	convertedTime := oldTime.In(t)

	parsedTime, err := time.Parse("2006-01-02T15:04:05-07:00",
		convertedTime.Format("2006-01-02T15:04:05-07:00"))
	if err != nil {
		log.Println("error when parse time", err)
		return ""
	}

	return parsedTime.Format("2006-01-02T15:04:05-07:00")
}

func GetCurrentTimeWithMillisFormatted() string {
	now := time.Now()
	baseFormatted := now.Format("20060102150405")
	millis := now.Nanosecond() / 1000000
	return fmt.Sprintf("%s%02d", baseFormatted, millis)
}

func DateStrMonthYear(date time.Time) string {
	date = date.In(loc)

	return fmt.Sprintf("%02d %s %d", date.Day(), MonthName(date), date.Year())
}

func DateStrMonthYearHour(date time.Time) string {
	date = date.In(loc)

	return fmt.Sprintf("%02d %s %d %s", date.Day(), MonthName(date), date.Year(), date.Format("03:04:05 PM"))
}

func MonthName(date time.Time) string {
	date = date.In(loc)

	return map[string]string{
		"1":  "January",
		"2":  "February",
		"3":  "March",
		"4":  "April",
		"5":  "May",
		"6":  "June",
		"7":  "July",
		"8":  "August",
		"9":  "September",
		"10": "October",
		"11": "November",
		"12": "December",
	}[fmt.Sprintf("%d", date.Month())]
}

func ConvertTimeFromTimeStringWithTimezone(timeStr string) (time.Time, *time.Location, error) {
	if !IsPatternMatch(`^([01]\d|2[0-3]):([0-5]\d):([0-5]\d) GMT([+-]\d{1,2})$`, timeStr) {
		return time.Time{}, nil, fmt.Errorf("invalid time regex")
	}

	// Split the time and timezone parts
	parts := strings.Split(timeStr, " ")
	if len(parts) != 2 {
		return time.Time{}, nil, fmt.Errorf("invalid time format")
	}

	// Parse the time part (20:00:00)
	parsedTime, err := time.Parse("15:04:05", parts[0])
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("error parsing time part: %v", err)
	}

	// Handle the timezone part (e.g., GMT+7 or GMT-5)
	timezonePart := strings.TrimPrefix(parts[1], "GMT")
	offsetHours, err := strconv.Atoi(timezonePart)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("error parsing timezone offset: %v", err)
	}

	// Create a location with the specified offset
	location := time.FixedZone(fmt.Sprintf("GMT%s", timezonePart), offsetHours*3600)

	// Combine the parsed time with the location's offset
	// Convert the parsedTime (without a date) into a full time object with the current date
	now := time.Now().UTC()
	resultTime := time.Date(now.Year(), now.Month(), now.Day(),
		parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, location)

	return resultTime, location, nil
}

func AddDaysSkippingWeekends(currentDate time.Time, days int) time.Time {
	// Loop to add days, skipping weekends
	for days > 0 {
		currentDate = currentDate.AddDate(0, 0, 1) // Add one day
		if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
			days--
		}
	}

	return currentDate
}

// GetCurrentDateOfLocation returns the current date in the specified location with the time set to midnight.
// The loc parameter specifies the desired time.Location.
// The returned time.Time value will have the year, month, and day set to the current date in the specified location,
// and the hour, minute, second, and nanosecond set to zero.
func GetCurrentDateOfLocation(loc *time.Location) time.Time {
	date := time.Now().In(loc)
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

// GetJakartaTimeLocation returns the time.Location for the "Asia/Jakarta" timezone.
func GetJakartaTimeLocation() *time.Location {
	loc, _ := time.LoadLocation(constant.TimeLoc)
	return loc
}

func ConvertToJakartaString(oldTime time.Time) string {
	loc := ConvertToJakarta(oldTime)
	return loc.Format("02/01/2006 15:04:05 MST")
}

// Datetime format yyyy-MM-dd HH:mm:ss. Make sure the values ​​are in the correct format as this function will not generate an error.
func ParseDatetimeToUTC(date string, loc *time.Location) time.Time {
	t, _ := time.ParseInLocation(time.DateTime, date, loc)
	return t.In(time.UTC)
}

func ParseISO8601DatetimeToUTC(date string) time.Time {
	t, _ := time.Parse(constant.ISO8601Datetime, date)
	return t.In(time.UTC)
}

func CalculateDaysBetween(t1, t2 time.Time) int {
	return int(math.Round((t2.Sub(t1).Hours() / 24)))
}

// TimeToUTC converts a time.Time to UTC based on the provided timezone.
// It takes a time.Time and a timezone string (e.g., "America/New_York", "Europe/London").
// If the timezone is invalid, it returns the original time and an error.
// Otherwise, it returns the time converted to UTC.
func TimeToUTC(t time.Time, timezone string) (time.Time, error) {
	if t.IsZero() {
		return t, fmt.Errorf("time is zero")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, fmt.Errorf("invalid Time-Zone format. Use valid Time-Zone")
	}

	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc).UTC(), nil
}

// GetTimeZoneFromContext retrieves a time zone location from the context.
// It extracts the time zone string from the context using constant.CtxTimeZone key and
// converts it into a time.Location object.
// Returns:
//   - *time.Location: The time zone location if successfully retrieved and loaded.
//   - error: An error if the context is nil, no time zone is found in the context,
//     or the time zone format is invalid.
func GetTimeZoneFromContext(ctx context.Context) (*time.Location, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	tz, ok := ctx.Value(constant.CtxTimeZone).(string)
	if !ok {
		return nil, fmt.Errorf("timezone not found in context")
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone format. Use valid timezone")
	}

	return loc, nil
}

func GetTimeLocationFromContext(ctx context.Context) *time.Location {
	loc, ok := ctx.Value(constant.CtxTimeLocation).(*time.Location)
	if !ok {
		loc, _ = time.LoadLocation(constant.TimeLoc)
	}
	return loc
}

func IsWithinLast24Hours(t time.Time) bool {
	return t.After(time.Now().Add(-24 * time.Hour))
}

func ParseTimeToDatetime(d time.Time, t string) (time.Time, error) {
	return time.ParseInLocation(time.DateTime, fmt.Sprintf("%s %s", d.Format(time.DateOnly), t), d.Location())
}
