package models

type Event string

const (
	StartEvent         Event = "start"
	UnknownEvent       Event = "unknown"
	CreateBalanceEvent Event = "balance/create"
)

var EventToFlow = map[Event]Flow{
	StartEvent:         StartFlow,
	CreateBalanceEvent: CreateBalanceFlow,
}
