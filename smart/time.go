package smart

import "time"

// MinutesFromYearStart returns the number of minutes from the start of t's year
// to t, in UTC. Used for holiday and summer period times.
func MinutesFromYearStart(t time.Time) int64 {
	utc := t.UTC()
	truncated := time.Date(utc.Year(), utc.Month(), utc.Day(), utc.Hour(), 0, 0, 0, time.UTC)
	yearStart := time.Date(utc.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	return int64(truncated.Sub(yearStart).Minutes())
}

// TimeFromYearMinutes converts minutes from year start back to a time.Time.
// The year parameter specifies which year to use.
func TimeFromYearMinutes(year int, minutes int64) time.Time {
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	return yearStart.Add(time.Duration(minutes) * time.Minute)
}

// DateToYearMinutes returns minutes from year start for a month/day combination.
// Uses current year (UTC). Month is 1-12, day is 1-31.
func DateToYearMinutes(month, day int) int64 {
	year := time.Now().UTC().Year()
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	return int64(date.Sub(yearStart).Minutes())
}

// YearMinutesToDate converts minutes from year start to month and day.
// Uses current year (UTC) for the conversion.
func YearMinutesToDate(minutes int64) (month, day int) {
	t := TimeFromYearMinutes(time.Now().UTC().Year(), minutes)
	return int(t.Month()), t.Day()
}

// WeekMinutes returns minutes from the start of the week (Monday 00:00).
// Used for weekly timer entries.
func WeekMinutes(weekday time.Weekday, hour, minute int) int {
	// Monday = 0 in FRITZ!Box API (Go: Monday = 1, Sunday = 0)
	var day int
	if weekday == time.Sunday {
		day = 6
	} else {
		day = int(weekday) - 1
	}
	return day*24*60 + hour*60 + minute
}

// ParseWeekMinutes converts minutes from week start to weekday, hour, minute.
func ParseWeekMinutes(minutes int) (weekday time.Weekday, hour, minute int) {
	day := minutes / (24 * 60)
	remainder := minutes % (24 * 60)
	hour = remainder / 60
	minute = remainder % 60

	if day == 6 {
		weekday = time.Sunday
	} else {
		weekday = time.Weekday(day + 1)
	}
	return
}
