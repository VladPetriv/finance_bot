package model

import (
	"fmt"
	"time"

	"github.com/lib/pq"
)

// AutomaticReport represents a job that contains a configuration for automatic report generation.
type AutomaticReport struct {
	ID          string         `db:"id"`
	BalanceIDs  pq.StringArray `db:"balance_ids"`
	CategoryIDs pq.StringArray `db:"category_ids"`

	Name   string                `db:"name"`
	Period AutomaticReportPeriod `db:"period"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetID returns the ID of the automatic report.
func (a AutomaticReport) GetID() string {
	return a.ID
}

// GetName returns the name of the automatic report.
func (a AutomaticReport) GetName() string {
	return a.Name
}

// GetDeletionMessage returns the deletion confirmation message for the automatic report.
func (a AutomaticReport) GetDeletionMessage() string {
	return fmt.Sprintf(
		"Are you sure you want to delete the automatic report '%s' (%s)?\nYou will no longer receive this report.",
		a.Name,
		a.Period.GetLabel(),
	)
}

// AutomaticReportPeriod represents the period for which automatic report will be generated and sent to the user.
type AutomaticReportPeriod string

const (
	// AutomaticReportPeriodDaily represents a daily report.
	AutomaticReportPeriodDaily AutomaticReportPeriod = "daily"
	// AutomaticReportPeriodWeekly represents a weekly report.
	AutomaticReportPeriodWeekly AutomaticReportPeriod = "weekly"
	// AutomaticReportPeriodMonthly represents a monthly report.
	AutomaticReportPeriodMonthly AutomaticReportPeriod = "monthly"
	// AutomaticReportPeriodQuarterly represents a quarterly report.
	AutomaticReportPeriodQuarterly AutomaticReportPeriod = "quarterly"
	// AutomaticReportPeriodYearly represents a yearly report.
	AutomaticReportPeriodYearly AutomaticReportPeriod = "yearly"
	// AutomaticReportPeriodEndOfDay represents a report for the calendar day that is sent at its end.
	AutomaticReportPeriodEndOfDay AutomaticReportPeriod = "end_of_day"
	// AutomaticReportPeriodEndOfWeek represents a report for the calendar week (Monday-Sunday) that is sent at its end.
	AutomaticReportPeriodEndOfWeek AutomaticReportPeriod = "end_of_week"
	// AutomaticReportPeriodEndOfMonth represents a report for the calendar month that is sent at its end.
	AutomaticReportPeriodEndOfMonth AutomaticReportPeriod = "end_of_month"
	// AutomaticReportPeriodEndOfQuarter represents a report for the calendar quarter that is sent at its end.
	AutomaticReportPeriodEndOfQuarter AutomaticReportPeriod = "end_of_quarter"
	// AutomaticReportPeriodEndOfYear represents a report for the calendar year that is sent at its end.
	AutomaticReportPeriodEndOfYear AutomaticReportPeriod = "end_of_year"
)

// ScheduledReportExecution represents a job that should be executed at a scheduled time with a specific report.
type ScheduledReportExecution struct {
	ID                string    `db:"id"`
	AutomaticReportID string    `db:"automatic_report_id"`
	ExecutionDate     time.Time `db:"execution_date"`
}

// ParseAutomaticReportPeriod parses a string into an AutomaticReportPeriod.
func ParseAutomaticReportPeriod(period string) (AutomaticReportPeriod, error) {
	switch period {
	case string(AutomaticReportPeriodDaily):
		return AutomaticReportPeriodDaily, nil
	case string(AutomaticReportPeriodWeekly):
		return AutomaticReportPeriodWeekly, nil
	case string(AutomaticReportPeriodMonthly):
		return AutomaticReportPeriodMonthly, nil
	case string(AutomaticReportPeriodQuarterly):
		return AutomaticReportPeriodQuarterly, nil
	case string(AutomaticReportPeriodYearly):
		return AutomaticReportPeriodYearly, nil
	case string(AutomaticReportPeriodEndOfDay):
		return AutomaticReportPeriodEndOfDay, nil
	case string(AutomaticReportPeriodEndOfWeek):
		return AutomaticReportPeriodEndOfWeek, nil
	case string(AutomaticReportPeriodEndOfMonth):
		return AutomaticReportPeriodEndOfMonth, nil
	case string(AutomaticReportPeriodEndOfQuarter):
		return AutomaticReportPeriodEndOfQuarter, nil
	case string(AutomaticReportPeriodEndOfYear):
		return AutomaticReportPeriodEndOfYear, nil
	default:
		return "", fmt.Errorf("invalid automatic report period: %s", period)
	}
}

// AddTo returns the date shifted forward by one period.
func (p AutomaticReportPeriod) AddTo(date time.Time) time.Time {
	switch p {
	case AutomaticReportPeriodDaily, AutomaticReportPeriodEndOfDay:
		return date.AddDate(0, 0, 1)
	case AutomaticReportPeriodWeekly, AutomaticReportPeriodEndOfWeek:
		return date.AddDate(0, 0, 7)
	case AutomaticReportPeriodMonthly, AutomaticReportPeriodEndOfMonth:
		return date.AddDate(0, 1, 0)
	case AutomaticReportPeriodQuarterly, AutomaticReportPeriodEndOfQuarter:
		return date.AddDate(0, 3, 0)
	case AutomaticReportPeriodYearly, AutomaticReportPeriodEndOfYear:
		return date.AddDate(1, 0, 0)
	default:
		return date
	}
}

// SubtractFrom returns the date shifted back by one period.
func (p AutomaticReportPeriod) SubtractFrom(date time.Time) time.Time {
	switch p {
	case AutomaticReportPeriodDaily, AutomaticReportPeriodEndOfDay:
		return date.AddDate(0, 0, -1)
	case AutomaticReportPeriodWeekly, AutomaticReportPeriodEndOfWeek:
		return date.AddDate(0, 0, -7)
	case AutomaticReportPeriodMonthly, AutomaticReportPeriodEndOfMonth:
		return date.AddDate(0, -1, 0)
	case AutomaticReportPeriodQuarterly, AutomaticReportPeriodEndOfQuarter:
		return date.AddDate(0, -3, 0)
	case AutomaticReportPeriodYearly, AutomaticReportPeriodEndOfYear:
		return date.AddDate(-1, 0, 0)
	default:
		return date
	}
}

// CalculateNextExecutionDate returns the first execution date after now, stepping forward by one period from the given date.
func (p AutomaticReportPeriod) CalculateNextExecutionDate(from, now time.Time) time.Time {
	next := p.AddTo(from)
	for !next.After(now) {
		next = p.AddTo(next)
	}

	return next
}

// CalculateFirstExecutionDate returns the date of the first report execution for a report created at the given time.
func (p AutomaticReportPeriod) CalculateFirstExecutionDate(now time.Time) time.Time {
	now = now.UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	switch p {
	case AutomaticReportPeriodEndOfWeek:
		daysUntilNextMonday := (int(time.Monday) - int(today.Weekday()) + 7) % 7
		if daysUntilNextMonday == 0 {
			daysUntilNextMonday = 7
		}

		return today.AddDate(0, 0, daysUntilNextMonday)
	case AutomaticReportPeriodEndOfMonth:
		return time.Date(today.Year(), today.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	case AutomaticReportPeriodEndOfQuarter:
		nextQuarterFirstMonth := (int(today.Month())-1)/3*3 + 4

		return time.Date(today.Year(), time.Month(nextQuarterFirstMonth), 1, 0, 0, 0, 0, time.UTC)
	case AutomaticReportPeriodEndOfYear:
		return time.Date(today.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	default:
		return p.CalculateNextExecutionDate(today, now)
	}
}

// GetLabel returns a human readable name of the period.
func (p AutomaticReportPeriod) GetLabel() string {
	switch p {
	case AutomaticReportPeriodDaily:
		return "Every day"
	case AutomaticReportPeriodWeekly:
		return "Every week"
	case AutomaticReportPeriodMonthly:
		return "Every month"
	case AutomaticReportPeriodQuarterly:
		return "Every quarter"
	case AutomaticReportPeriodYearly:
		return "Every year"
	case AutomaticReportPeriodEndOfDay:
		return "End of day"
	case AutomaticReportPeriodEndOfWeek:
		return "End of week"
	case AutomaticReportPeriodEndOfMonth:
		return "End of month"
	case AutomaticReportPeriodEndOfQuarter:
		return "End of quarter"
	case AutomaticReportPeriodEndOfYear:
		return "End of year"
	default:
		return string(p)
	}
}
