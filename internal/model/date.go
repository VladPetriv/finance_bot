package model

import "time"

// Month is a type representing a month of the year
type Month string

const (
	// MonthJanuary represent January month
	MonthJanuary Month = "January"
	// MonthFebruary represent February month
	MonthFebruary Month = "February"
	// MonthMarch represent March month
	MonthMarch Month = "March"
	// MonthApril represent April month
	MonthApril Month = "April"
	// MonthMay represent May month
	MonthMay Month = "May"
	// MonthJune represent June month
	MonthJune Month = "June"
	// MonthJuly represent July month
	MonthJuly Month = "July"
	// MonthAugust represent August month
	MonthAugust Month = "August"
	// MonthSeptember represent September month
	MonthSeptember Month = "September"
	// MonthOctober represent October month
	MonthOctober Month = "October"
	// MonthNovember represent November month
	MonthNovember Month = "November"
	// MonthDecember represent December month
	MonthDecember Month = "December"
)

// GetID returns the string representation of the month
func (m Month) GetID() string {
	return string(m)
}

// GetName returns the string representation of the month
func (m Month) GetName() string {
	return string(m)
}

// GetIndex returns the index of the month
func (m Month) GetIndex() int {
	return monthToMonthIndex[m]
}

var monthToMonthIndex = map[Month]int{
	MonthJanuary:   1,
	MonthFebruary:  2,
	MonthMarch:     3,
	MonthApril:     4,
	MonthMay:       5,
	MonthJune:      6,
	MonthJuly:      7,
	MonthAugust:    8,
	MonthSeptember: 9,
	MonthOctober:   10,
	MonthNovember:  11,
	MonthDecember:  12,
}

// GetTimeRange returns the start and end time of the month.
// If the current time is in the same month, the start time is the current time and the end time is the end of the day.
func (m Month) GetTimeRange(currentTime time.Time) (time.Time, time.Time) {
	selectedMonth := time.Month(m.GetIndex())

	if currentTime.Month() == selectedMonth {
		startTime := time.Date(currentTime.Year(), selectedMonth, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(
			currentTime.Year(),
			selectedMonth,
			currentTime.Day(),
			currentTime.Hour(),
			currentTime.Minute(),
			currentTime.Second(),
			0,
			time.UTC,
		)

		return startTime, endTime
	}

	startTime := time.Date(currentTime.Year(), selectedMonth, 1, 0, 0, 0, 0, time.UTC)

	// Get the first day of next month
	nextMonth := selectedMonth + 1
	nextYear := currentTime.Year()

	// Handle December -> January transition
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}

	// Set endTime to the last day of the month (by taking first day of next month and subtracting 1 day)
	firstDayOfNextMonth := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, time.UTC)
	endTime := firstDayOfNextMonth.AddDate(0, 0, -1)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.UTC)

	return startTime, endTime
}

// Months represents an array of all months
var Months = []Month{
	MonthJanuary, MonthFebruary, MonthMarch,
	MonthApril, MonthMay, MonthJune,
	MonthJuly, MonthAugust, MonthSeptember,
	MonthOctober, MonthNovember, MonthDecember,
}

// CreationPeriod defines constants representing different time periods for creation operations.
type CreationPeriod string

const (
	// CreationPeriodDay represents a daily time period.
	CreationPeriodDay CreationPeriod = "day"
	// CreationPeriodWeek represents a weekly time period.
	CreationPeriodWeek CreationPeriod = "week"
	// CreationPeriodMonth represents a monthly time period.
	CreationPeriodMonth CreationPeriod = "month"
	// CreationPeriodYear represents a yearly time period.
	CreationPeriodYear CreationPeriod = "year"
	// CreationPeriodCurrentMonth represents the current month.
	CreationPeriodCurrentMonth CreationPeriod = "current_month"
)

// CalculateTimeRange is used to calculate start and end times based on a given period
func (c CreationPeriod) CalculateTimeRange() (time.Time, time.Time) {
	now := time.Now()

	startTime := now
	endTime := now

	switch c {
	case CreationPeriodDay:
		startTime = now.Add(-24 * time.Hour)
	case CreationPeriodWeek:
		startTime = now.Add(-7 * 24 * time.Hour)
	case CreationPeriodMonth:
		startTime = now.Add(-30 * 24 * time.Hour)
	case CreationPeriodYear:
		startTime = now.Add(-365 * 24 * time.Hour)
	case CreationPeriodCurrentMonth:
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	return startTime, endTime
}

// GetCreationPeriodFromText checks if the input text matches any of the CreationPeriod enums.
// If there is no match, it returns nil. Otherwise, it returns the corresponding CreationPeriod type.
func GetCreationPeriodFromText(value string) CreationPeriod {
	switch value {
	case string(CreationPeriodDay):
		return CreationPeriodDay
	case string(CreationPeriodWeek):
		return CreationPeriodWeek
	case string(CreationPeriodMonth):
		return CreationPeriodMonth
	case string(CreationPeriodYear):
		return CreationPeriodYear
	default:
		return ""
	}
}
