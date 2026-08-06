package types_test

import (
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	require.NotNil(t, loc)
	require.NoError(t, err)

	toTime := func(t time.Time) Time {
		return Time{Time: t}
	}

	dt := Time{Time: time.Date(2025, 10, 22, 17, 30, 1, 0, loc)}
	assert.Equal(t, toTime(time.Date(2025, 10, 23, 0, 0, 0, 0, loc)), dt.NextMidnight())
	assert.Equal(t, toTime(time.Date(2025, 10, 24, 0, 0, 0, 0, loc)), dt.NextMidnight().NextMidnight())
	assert.Equal(t, toTime(time.Date(2025, 10, 31, 0, 0, 0, 0, loc)), dt.EndOfMonth())
	assert.Equal(t, toTime(time.Date(2025, 11, 01, 0, 0, 0, 0, loc)), dt.EndOfMonth().NextMidnight())
	assert.Equal(t, 4, dt.Quarter())
	assert.Equal(t, false, dt.IsEndOfQuerter())
	assert.Equal(t, 3, toTime(dt.AddDate(0, -1, 0)).Quarter())
	assert.Equal(t, true, toTime(dt.AddDate(0, -1, 0)).IsEndOfQuerter())
	assert.Equal(t, 3, toTime(dt.AddDate(0, -2, 0)).Quarter())
	assert.Equal(t, 2, toTime(dt.AddDate(0, -4, 0)).Quarter())
	assert.Equal(t, 1, toTime(dt.AddDate(0, -9, 0)).Quarter())
}
