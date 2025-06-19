package common

import "time"

var JakartaTZ = mustLoadAsiaJakarta()

func mustLoadAsiaJakarta() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic("failed to load Asia/Jakarta timezone: " + err.Error())
	}
	return loc
}

func TruncateToJakartaDate(t time.Time) time.Time {
	t = t.In(JakartaTZ)
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, JakartaTZ)
}

// NewDate returns time at 00:00:00 in UTC+7 (Asia/Jakarta)
func NewDate(year int, month int, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, JakartaTZ)
}

// NewDateTime returns time at given hour/minute/second in UTC+7
func NewDateTime(year, month, day, hour, min, sec int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, JakartaTZ)
}

// NewDateToday returns today's date at 00:00:00 in Asia/Jakarta timezone
func NewDateToday() time.Time {
	now := time.Now().In(JakartaTZ)
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, JakartaTZ)
}

// NewDateTimeNow returns current date and time in Asia/Jakarta timezone
func NewDateTimeNow() time.Time {
	return time.Now().In(JakartaTZ)
}
