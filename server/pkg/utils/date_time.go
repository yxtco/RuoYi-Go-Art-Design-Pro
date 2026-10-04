package utils

import "time"

func GetCurrentDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")

}

func GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

func ParseDateTime(dateTime string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", dateTime, time.Local)
}

func ParseDate(date string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", date, time.Local)
}

func ParseDateMonth(date string) (time.Time, error) {
	return time.ParseInLocation("2006-01", date, time.Local)
}

func ParseTimeToDate(timestamp int64) string {
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02")
}
