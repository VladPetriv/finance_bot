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

	if s.Flow == CancelFlow && len(s.Steps) == 1 {
		return CancelEvent
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
	case DeleteBalanceFlowStep:
		return DeleteBalanceEvent
	case CreateCategoryFlowStep:
		return CreateCategoryEvent
	case ListCategoriesFlowStep:
		return ListCategoriesEvent
	case UpdateCategoryFlowStep:
		return UpdateCategoryEvent
	case DeleteCategoryFlowStep:
		return DeleteCategoryEvent
	case CreateOperationFlowStep:
		return CreateOperationEvent
	case GetOperationsHistoryFlowStep:
		return GetOperationsHistoryEvent
	default:
		return UnknownEvent
	}
}

// Flow represents the type of interaction flow currently active
type Flow string

const (
	// StartFlow represents the initial flow when starting the bot
	StartFlow Flow = "start"
	// CancelFlow represents the flow for stopping current flow
	CancelFlow Flow = "cancel"

	// BalanceFlow represents the flow for getting balance actions
	BalanceFlow Flow = "balance"
	// CategoryFlow represents the flow for getting category actions
	CategoryFlow Flow = "category"
	// OperationFlow represents the flow for getting operation actions
	OperationFlow Flow = "operation"

	// CreateBalanceFlow represents the flow for creating a new balance
	CreateBalanceFlow Flow = "create_balance"
	// UpdateBalanceFlow represents the flow for updating a balance
	UpdateBalanceFlow Flow = "update_balance"
	// GetBalanceFlow represents the flow for getting a balance
	GetBalanceFlow Flow = "get_balance"
	// DeleteBalanceFlow represents the flow for deleting a balance
	DeleteBalanceFlow Flow = "delete_balance"

	// CreateCategoryFlow represents the flow for creating a new category
	CreateCategoryFlow Flow = "create_category"
	// ListCategoriesFlow represents the flow for listing all categories
	ListCategoriesFlow Flow = "list_categories"
	// UpdateCategoryFlow represents the flow for updating a category
	UpdateCategoryFlow Flow = "update_category"
	// DeleteCategoryFlow represents the flow for deleting a category
	DeleteCategoryFlow Flow = "delete_category"

	// CreateOperationFlow represents the flow for creating a new operation
	CreateOperationFlow Flow = "create_operation"
	// GetOperationsHistoryFlow represents the flow for getting operations history
	GetOperationsHistoryFlow Flow = "get_operations_history"
)

// FlowStep represents a specific step within a flow
type FlowStep string

const (
	// General steps

	// StartFlowStep represents the initial step of any flow
	StartFlowStep FlowStep = "start"
	// EndFlowStep represents the final step of any flow
	EndFlowStep FlowStep = "end"

	// Steps that are related for balance

	// CreateInitialBalanceFlowStep represents the step for creating the first balance
	CreateInitialBalanceFlowStep FlowStep = "create_initial_balance"
	// CreateBalanceFlowStep represents the step for creating additional balances
	CreateBalanceFlowStep FlowStep = "create_balance"
	// UpdateBalanceFlowStep represents the step for updating a balance
	UpdateBalanceFlowStep FlowStep = "update_balance"
	// GetBalanceFlowStep represents the step for getting a balance
	GetBalanceFlowStep FlowStep = "get_balance"
	// DeleteBalanceFlowStep represents the step for deleting a balance
	DeleteBalanceFlowStep FlowStep = "delete_balance"
	// ChooseBalanceFlowStep represents the step for choosing balance that will be used for an action
	ChooseBalanceFlowStep FlowStep = "choose_balance"
	// ConfirmBalanceDeletionFlowStep represents the step for confirming balance deletion
	ConfirmBalanceDeletionFlowStep FlowStep = "confirm_balance_deletion"
	// EnterBalanceNameFlowStep represents the step for entering balance name
	EnterBalanceNameFlowStep FlowStep = "enter_balance_name"
	// EnterBalanceCurrencyFlowStep represents the step for entering balance currency
	EnterBalanceCurrencyFlowStep FlowStep = "enter_balance_currency"
	// EnterBalanceAmountFlowStep represents the step for entering balance amount
	EnterBalanceAmountFlowStep FlowStep = "enter_balance_amount"

	// Steps that are related for category

	// CreateCategoryFlowStep represents the step for creating a new category
	CreateCategoryFlowStep FlowStep = "create_category"
	// UpdateCategoryFlowStep represents the step for updating a category
	UpdateCategoryFlowStep FlowStep = "update_category"
	// DeleteCategoryFlowStep represents the step for deleting a category
	DeleteCategoryFlowStep FlowStep = "delete_category"
	// ChooseCategoryFlowStep represents the step for choosing category
	ChooseCategoryFlowStep FlowStep = "choose_category"
	// EnterUpdatedCategoryNameFlowStep represents the step for entering category name
	EnterUpdatedCategoryNameFlowStep FlowStep = "enter_updated_category_name"
	// EnterCategoryNameFlowStep represents the step for entering category name
	EnterCategoryNameFlowStep FlowStep = "enter_category_name"
	// ListCategoriesFlowStep represents the step for listing all categories
	ListCategoriesFlowStep FlowStep = "list_categories"

	// Steps that are related for operation

	// CreateOperationFlowStep represents the step for creating a new operation
	CreateOperationFlowStep FlowStep = "create_operation"
	// ProcessOperationTypeFlowStep represents the step for processing operation type
	ProcessOperationTypeFlowStep FlowStep = "process_opration_type"
	// ChooseBalanceFromFlowStep represents the step for choosing balance from which transfer operation will be created
	ChooseBalanceFromFlowStep FlowStep = "choose_balance_from_for_transfer_operation"
	// ChooseBalanceToFlowStep represents the step for choosing balance to which transfer operation will be created
	ChooseBalanceToFlowStep FlowStep = "choose_balance_to_for_transfer_operation"
	// EnterCurrencyExchangeRateFlowStep represents the step for entering currency exchange rate
	EnterCurrencyExchangeRateFlowStep FlowStep = "enter_currency_exchange_rate"
	// EnterOperationDescriptionFlowStep represents the step for entering operation description
	EnterOperationDescriptionFlowStep FlowStep = "enter_operation_description"
	// EnterOperationAmountFlowStep represents the step for entering operation amount
	EnterOperationAmountFlowStep FlowStep = "enter_operation_amount"
	// GetOperationsHistoryFlowStep represents the step for getting operations history
	GetOperationsHistoryFlowStep FlowStep = "get_operations_history"
	// ChooseTimePeriodForOperationsHistoryFlowStep represents the step for choosing time period for operations history
	ChooseTimePeriodForOperationsHistoryFlowStep FlowStep = "choose_time_period_for_operations_history"
)
