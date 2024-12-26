package models

import (
	"time"
)

// State represents the current state of a user's interaction with the bot
type State struct {
	ID     string `bson:"_id"`
	UserID string `bson:"userId"`

	Flow  Flow       `bson:"flow"`
	Steps []FlowStep `bson:"steps"`

	Metedata map[string]any `bson:"metadata"`

	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

// GetCurrentStep returns the current step in the flow
func (s *State) GetCurrentStep() FlowStep {
	return s.Steps[len(s.Steps)-1]
}

// IsFlowFinished checks if the current flow has reached its end
func (s *State) IsFlowFinished() bool {
	return s.Steps[len(s.Steps)-1] == EndFlowStep
}

const indexOfInitialFlowStep = 1

// GetEvent determines the current event based on the flow state
func (s *State) GetEvent() Event {
	if s.Flow == CreateBalanceFlow && len(s.Steps) == 1 {
		return CreateBalanceEvent
	}

	switch s.Steps[indexOfInitialFlowStep] {
	case CreateInitialBalanceFlowStep:
		return CreateBalanceEvent
	case CreateBalanceFlowStep:
		return CreateBalanceEvent
	case UpdateBalanceFlowStep:
		return UpdateBalanceEvent
	case GetBalanceFlowStep:
		return GetBalanceEvent
	case CreateCategoryFlowStep:
		return CreateCategoryEvent
	case ListCategoriesFlowStep:
		return ListCategoriesEvent
	case CreateOperationFlowStep:
		return CreateOperationEvent
	default:
		return UnknownEvent
	}
}

// Flow represents the type of interaction flow currently active
type Flow string

const (
	// StartFlow represents the initial flow when starting the bot
	StartFlow Flow = "start"
	// CreateBalanceFlow represents the flow for creating a new balance
	CreateBalanceFlow Flow = "create_balance"
	// UpdateBalanceFlow represents the flow for updating a balance
	UpdateBalanceFlow Flow = "update_balance"
	// GetBalanceFlow represents the flow for getting a balance
	GetBalanceFlow Flow = "get_balance"
	// CreateCategoryFlow represents the flow for creating a new category
	CreateCategoryFlow Flow = "create_category"
	// ListCategoriesFlow represents the flow for listing all categories
	ListCategoriesFlow Flow = "list_categories"
	// CreateOperationFlow represents the flow for creating a new operation
	CreateOperationFlow Flow = "create_operation"
)

// FlowStep represents a specific step within a flow
type FlowStep string

const (
	// General steps

	// StartFlowStep represents the initial step of any flow
	StartFlowStep FlowStep = "start"
	// EndFlowStep represents the final step of any flow
	EndFlowStep FlowStep = "end"

	// Steps that are relad for balance

	// CreateInitialBalanceFlowStep represents the step for creating the first balance
	CreateInitialBalanceFlowStep FlowStep = "create_initial_balance"
	// CreateBalanceFlowStep represents the step for creating additional balances
	CreateBalanceFlowStep FlowStep = "create_balance"
	// UpdateBalanceFlowStep represents the step for updating a balance
	UpdateBalanceFlowStep FlowStep = "update_balance"
	// GetBalanceFlowStep represents the step for getting a balance
	GetBalanceFlowStep FlowStep = "get_balance"
	// ChooseBalanceFlowStep represents the step for choosing balance that will be used for an action
	ChooseBalanceFlowStep FlowStep = "choose_balance"
	// EnterBalanceNameFlowStep represents the step for entering balance name
	EnterBalanceNameFlowStep FlowStep = "enter_balance_name"
	// EnterBalanceCurrencyFlowStep represents the step for entering balance currency
	EnterBalanceCurrencyFlowStep FlowStep = "enter_balance_currency"
	// EnterBalanceAmountFlowStep represents the step for entering balance amount
	EnterBalanceAmountFlowStep FlowStep = "enter_balance_amount"

	// Steps that are relad for category

	// CreateCategoryFlowStep represents the step for creating a new category
	CreateCategoryFlowStep FlowStep = "create_category"
	// EnterCategoryNameFlowStep represents the step for entering category name
	EnterCategoryNameFlowStep FlowStep = "enter_category_name"
	// ListCategoriesFlowStep represents the step for listing all categories
	ListCategoriesFlowStep FlowStep = "list_categories"

	// Steps that are relad for operation

	// CreateOperationFlowStep represents the step for creating a new operation
	CreateOperationFlowStep FlowStep = "create_operation"
	// ChooseOprationTypeFlowStep represents the step for choosing operation type
	ChooseOprationTypeFlowStep FlowStep = "choose_opration_type"
	// ChooseCategoryFlowStep represents the step for choosing category
	ChooseCategoryFlowStep FlowStep = "choose_category"
	// EnterOperationAmountFlowStep represents the step for entering operation amount
	EnterOperationAmountFlowStep FlowStep = "enter_operation_amount"
)
