package utils

import "time"

func ParseDate(date string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return time.Time{}, err
	}

	return parsed, nil
}

func FormatDate(date *time.Time) string {
	return date.Format(time.RFC3339)
}

func FormatDateToRFC3339TimeZone(date *time.Time, timeZone string) *string {
	location, _ := time.LoadLocation(timeZone)
	inTimeZone := date.In(location)
	birthdayWithTimeZone := FormatDate(&inTimeZone)
	return &birthdayWithTimeZone
}

func GetDefaultTimeZone() *time.Location {
	location, err := time.LoadLocation("Asia/Tbilisi")
	if err != nil {
		return time.UTC
	}
	return location
}
