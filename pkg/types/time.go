package types

import "time"

type Time struct {
	time.Time
}

func (t Time) NextMidnight() Time {
	return Time{Time: time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())}
}

func (t Time) EndOfMonth() Time {
	return Time{Time: time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location()).AddDate(0, 0, -1)}
}

func (t Time) Quarter() int {
	if m := t.Month(); m <= time.March {
		return 1

	} else if m <= time.June {
		return 2

	} else if m <= time.September {
		return 3

	} else {
		return 4
	}
}

func (t Time) IsEndOfQuerter() bool {
	switch t.Month() {
	default:
		return false

	case time.March, time.June, time.September, time.December:
		return true
	}
}
