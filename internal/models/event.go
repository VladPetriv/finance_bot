package models

// Event represents the type of event that occurs during bot interaction
type Event string

const (
	// StartEvent represents the event when user starts interacting with the bot
	StartEvent Event = "start"
	// UnknownEvent represents an unrecognized or unsupported event
	UnknownEvent Event = "unknown"
	// CreateBalanceEvent represents the event for creating a new balance
	CreateBalanceEvent Event = "balance/create"
	// UpdateBalanceEvent represents the event for updating a balance
	UpdateBalanceEvent Event = "balance/update"
	// GetBalanceEvent represents the event for getting a balance
	GetBalanceEvent Event = "balance/get"
)

// EventToFlow maps events to their corresponding flows
var EventToFlow = map[Event]Flow{
	StartEvent:         StartFlow,
	CreateBalanceEvent: CreateBalanceFlow,
	UpdateBalanceEvent: UpdateBalanceFlow,
	GetBalanceEvent:    GetBalanceFlow,
}
