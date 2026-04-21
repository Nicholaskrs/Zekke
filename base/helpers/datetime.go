package helpers

import (
	"fmt"
	"strconv"
	"time"
)

const (
	FORMAT_DB_DATE_TIME          = "2006-01-02T15:04:05Z"
	FORMAT_DAY_FIRST             = "02-01-2006"
	FORMAT_DAY_FIRST_SLASH       = "02/01/2006"
	FORMAT_DATE                  = "2006-01-02"
	FORMAT_DATETIME              = "2006-01-02 15:04:05"
	FORMAT_DATE_TIME_FIRST       = "02-01-2006 15:04:05"
	FORMAT_DATETIME_NO_SEPARATOR = "20060102150405"
)

func FormatDateTime(dateTime string, layoutFrom string, layoutTo string) string {
	date, _ := time.Parse(layoutFrom, dateTime)
	return date.Format(layoutTo)
}

func FormatToDateTime(dateString string, layout string) (time.Time, error) {
	// Parse the input date string using the input format
	date, err := time.Parse(layout, dateString)
	if err != nil {
		date, _ := time.Parse("2006-01-01", "0000-00-00")
		return date, err
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	return date, nil
}

// FormatDatetimeToString used to change from type time.Time to string in format using layout from parameter
func FormatDatetimeToString(currentTime time.Time, layoutTo string) string {
	formattedDateTime := currentTime.Format(layoutTo)

	return formattedDateTime
}

// ExcelDateToTime returns formatted Excel raw date to time.Time format
// Excel date system starts from January 1, 1900, which is serial 1
// In the serial date system, 0 corresponds to December 30, 1899.
func ExcelDateToTime(serial string) (time.Time, error) {
	excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	serialInt, err := strconv.Atoi(serial)
	if err != nil {
		parsedTime, err := FormatToDateTime(serial, FORMAT_DAY_FIRST_SLASH)
		if err != nil {
			return time.Time{}, err
		}
		return parsedTime, nil
	}
	return excelEpoch.AddDate(0, 0, serialInt), nil
}

func DateOnlyFormat(dateTime time.Time) time.Time {
	return time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, time.Local)
}

func CalculateFuncTimeExecution(f func()) {
	startTime := time.Now()
	f()
	elapsedTime := time.Since(startTime) // Calculate elapsed time
	fmt.Printf("Function execution time: %v ms\n", elapsedTime.Milliseconds())
}
