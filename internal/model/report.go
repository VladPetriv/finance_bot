package model

import (
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
)

// ScheduledReportExecution represents a job that should be executed at a scheduled time with a specific report.
type ScheduledReportExecution struct {
	ID                string    `db:"id"`
	AutomaticReportID string    `db:"automatic_report_id"`
	ExecutionDate     time.Time `db:"execution_date"`
}
