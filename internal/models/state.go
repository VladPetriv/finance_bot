package models

import (
	"time"
)

type State struct {
	ID     string `bson:"_id"`
	UserID string `bson:"userId"`

	Flow  Flow       `bson:"flow"`
	Steps []FlowStep `bson:"steps"`

	Metedata map[string]any `bson:"metadata"`

	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func (s *State) GetCurrentStep() FlowStep {
	return s.Steps[len(s.Steps)-1]
}

func (s *State) IsFlowFinished() bool {
	return s.Steps[len(s.Steps)-1] == EndFlowStep
}

const indexOfInitialFlowStep = 1

func (s *State) GetEvent() Event {
	if len(s.Steps) < 2 {
		return UnknownEvent
	}

	if s.Flow == CreateBalanceFlow && len(s.Steps) == 1 {
		return CreateBalanceEvent
	}

	switch s.Steps[indexOfInitialFlowStep] {
	case CreateInitialBalanceFlowStep:
		return CreateBalanceEvent

	default:
		return UnknownEvent
	}
}

type Flow string

const (
	StartFlow         Flow = "start"
	CreateBalanceFlow Flow = "create_balance"
)

type FlowStep string

const (
	// Common flow steps
	StartFlowStep FlowStep = "start"
	EndFlowStep   FlowStep = "end"

	// Balance flow steps
	CreateInitialBalanceFlowStep FlowStep = "create_initial_balance"
	CreateBalanceFlowStep        FlowStep = "create_balance"
	EnterBalanceNameFlowStep     FlowStep = "enter_balance_name"
	EnterBalanceCurrencyFlowStep FlowStep = "enter_balance_currency"
	EnterBalanceAmountFlowStep   FlowStep = "enter_balance_amount"
)
