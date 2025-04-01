package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// State represents the current state of a user's interaction with the bot
type State struct {
	ID     string `db:"id"`
	UserID string `db:"user_username"`

	Flow  Flow      `db:"flow"`
	Steps FlowSteps `db:"steps"`

	Metedata Metadata `db:"metadata"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetFlowName returns the flow name in pretty format.
func (s *State) GetFlowName() string {
	parts := strings.Split(string(s.Flow), "_")

	var result string
	for index, part := range parts {
		if index == 0 {
			caser := cases.Title(language.English)
			result += caser.String(part)

			continue
		}

		result += " " + part
	}

	return result
}

// GetCurrentStep returns the current step in the flow
func (s *State) GetCurrentStep() FlowStep {
	return s.Steps[len(s.Steps)-1]
}

// IsFlowFinished checks if the current flow has reached its end
func (s *State) IsFlowFinished() bool {
	return s.Steps[len(s.Steps)-1] == EndFlowStep
}

// IsCommandAllowedDuringFlow checks if the command is allowed during the current flow
func (s *State) IsCommandAllowedDuringFlow(command string) bool {
	switch s.Flow {
	case CreateOperationFlow:
		if s.GetCurrentStep() == ProcessOperationTypeFlowStep {
			return slices.Contains(
				[]string{BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand, BotCreateTransferOperationCommand},
				command,
			)
		}

		return false

	case DeleteOperationFlow:
		if s.GetCurrentStep() == ChooseOperationToDeleteFlowStep {
			return slices.Contains(
				[]string{BotShowMoreCommand},
				command,
			)
		}

		return false

	case UpdateOperationFlow:
		switch s.GetCurrentStep() {
		case ChooseOperationToUpdateFlowStep:
			return slices.Contains(
				[]string{BotShowMoreCommand},
				command,
			)
		case ChooseUpdateOperationOptionFlowStep:
			return slices.Contains(
				[]string{
					BotUpdateOperationAmountCommand, BotUpdateOperationDescriptionCommand,
					BotUpdateOperationCategoryCommand, BotUpdateOperationDateCommand,
				},
				command,
			)
		}

		return false

	case UpdateBalanceFlow:
		switch s.GetCurrentStep() {
		case ChooseUpdateBalanceOptionFlowStep:
			return slices.Contains(
				[]string{
					BotUpdateBalanceNameCommand, BotUpdateBalanceAmountCommand, BotUpdateBalanceCurrencyCommand,
				},
				command,
			)
		}

		return false

	case UpdateBalanceSubscriptionFlow:
		switch s.GetCurrentStep() {
		case ChooseBalanceSubscriptionToUpdateFlowStep:
			return slices.Contains(
				[]string{
					BotShowMoreCommand,
				},
				command,
			)

		case ChooseUpdateBalanceSubscriptionOptionFlowStep:
			return slices.Contains(
				[]string{
					BotUpdateBalanceSubscriptionNameCommand, BotUpdateBalanceSubscriptionAmountCommand,
					BotUpdateBalanceSubscriptionCategoryCommand, BotUpdateBalanceSubscriptionPeriodCommand,
				},
				command,
			)
		}

		return false

	default:
		return false
	}
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
	case CreateInitialBalanceFlowStep, CreateBalanceFlowStep:
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
	case DeleteOperationFlowStep:
		return DeleteOperationEvent
	case GetOperationsHistoryFlowStep:
		return GetOperationsHistoryEvent
	case UpdateOperationFlowStep:
		return UpdateOperationEvent
	case CreateOperationsThroughOneTimeInputFlowStep:
		return CreateOperationsThroughOneTimeInputEvent
	case CreateBalanceSubscriptionFlowStep:
		return CreateBalanceSubscriptionEvent
	case ListBalanceSubscriptionFlowStep:
		return ListBalanceSubscriptionEvent
	case UpdateBalanceSubscriptionFlowStep:
		return UpdateBalanceSubscriptionEvent
	case DeleteBalanceSubscriptionFlowStep:
		return DeleteBalanceSubscriptionEvent
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
	// BalanceSubscriptionFlow represents the flow for getting balance subscriptions actions
	BalanceSubscriptionFlow Flow = "balance_subscriptions"

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
	// DeleteOperationFlow represents the flow for deleting an operation
	DeleteOperationFlow Flow = "delete_operation"
	// UpdateOperationFlow represents the flow for updating an operation
	UpdateOperationFlow Flow = "update_operation"
	// CreateOperationsThroughOneTimeInputFlow represents the flow for creating operations through one-time input
	CreateOperationsThroughOneTimeInputFlow Flow = "create_operations_through_one_time_input"

	// CreateBalanceSubscriptionFlow represents the flow for creating a new balance subscription
	CreateBalanceSubscriptionFlow Flow = "create_balance_subscription"
	// ListBalanceSubscriptionFlow represents the flow for listing all balance subscriptions
	ListBalanceSubscriptionFlow Flow = "list_balance_subscriptions"
	// UpdateBalanceSubscriptionFlow represents the flow for updating a balance subscription
	UpdateBalanceSubscriptionFlow Flow = "update_balance_subscription"
	// DeleteBalanceSubscriptionFlow represents the flow for deleting a balance subscription
	DeleteBalanceSubscriptionFlow Flow = "delete_balance_subscription"
)

// FlowSteps represents a slice of FlowStep
type FlowSteps []FlowStep

// Value implements the driver.Valuer interface
func (s FlowSteps) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface
func (s *FlowSteps) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, s)
}

// Metadata represents metadata associated with a flow
type Metadata map[string]any

// Value implements the driver.Valuer interface
func (s Metadata) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface
func (s *Metadata) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, s)
}

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
	// ChooseUpdateBalanceOptionFlowStep represents the step for choosing update balance option
	ChooseUpdateBalanceOptionFlowStep FlowStep = "choose_update_balance_option"
	// GetBalanceFlowStep represents the step for getting a balance
	GetBalanceFlowStep FlowStep = "get_balance"
	// ChooseMonthBalanceStatisticsFlowStep represents the step for choosing month for balance statistics
	ChooseMonthBalanceStatisticsFlowStep FlowStep = "choose_month_for_balance_statistics"
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
	ProcessOperationTypeFlowStep FlowStep = "process_operation_type"
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
	// DeleteOperationFlowStep represents the step for deleting an operation
	DeleteOperationFlowStep FlowStep = "delete_operation"
	// ChooseOperationToDeleteFlowStep represents the step for choosing operation to delete
	ChooseOperationToDeleteFlowStep FlowStep = "choose_operation_to_delete"
	// ConfirmOperationDeletionFlowStep represents the step for confirming operation deletion
	ConfirmOperationDeletionFlowStep FlowStep = "confirm_operation_deletion"
	// UpdateOperationFlowStep represents the step for updating an operation
	UpdateOperationFlowStep FlowStep = "update_operation"
	// ChooseOperationToUpdateFlowStep represents the step for choosing operation to update
	ChooseOperationToUpdateFlowStep FlowStep = "choose_operation_to_update"
	// ChooseUpdateOperationOptionFlowStep represents the step for choosing update operation option
	ChooseUpdateOperationOptionFlowStep FlowStep = "choose_update_operation_option"
	// EnterOperationDateFlowStep represents the step for entering operation date
	EnterOperationDateFlowStep FlowStep = "enter_operation_date"
	// CreateOperationsThroughOneTimeInputFlowStep represents the step for creating operations through one-time input
	CreateOperationsThroughOneTimeInputFlowStep FlowStep = "create_operations_through_one_time_input"
	// ConfirmOperationDetailsFlowStep represents the step for confirming operation details
	ConfirmOperationDetailsFlowStep FlowStep = "confirm_operation_details"

	// Steps that are related for balance subscription

	// CreateBalanceSubscriptionFlowStep represents the step for creating a balance subscription
	CreateBalanceSubscriptionFlowStep FlowStep = "create_balance_subscription"
	// EnterBalanceSubscriptionNameFlowStep represents the step for entering balance subscription name
	EnterBalanceSubscriptionNameFlowStep FlowStep = "enter_balance_subscription_name"
	// EnterBalanceSubscriptionAmountFlowStep represents the step for entering balance subscription amount
	EnterBalanceSubscriptionAmountFlowStep FlowStep = "enter_balance_subscription_amount"
	// ChooseBalanceSubscriptionFrequencyFlowStep represents the step for choosing balance subscription frequency
	ChooseBalanceSubscriptionFrequencyFlowStep FlowStep = "choose_balance_subscription_frequency"
	// EnterStartAtDateForBalanceSubscriptionFlowStep represents the step for entering start at date for balance subscription
	EnterStartAtDateForBalanceSubscriptionFlowStep FlowStep = "enter_start_at_date_for_balance_subscription"
	// ListBalanceSubscriptionFlowStep represents the step for listing balance subscriptions
	ListBalanceSubscriptionFlowStep FlowStep = "list_balance_subscriptions"
	// UpdateBalanceSubscriptionFlowStep represents the step for updating a balance subscription
	UpdateBalanceSubscriptionFlowStep FlowStep = "update_balance_subscription"
	// ChooseBalanceSubscriptionToUpdateFlowStep represents the step for choosing balance subscription to update
	ChooseBalanceSubscriptionToUpdateFlowStep FlowStep = "choose_balance_subscription_to_update"
	// ChooseUpdateBalanceSubscriptionOptionFlowStep represents the step for choosing balance subscription option
	ChooseUpdateBalanceSubscriptionOptionFlowStep FlowStep = "choose_update_balance_subscription_option"
	// DeleteBalanceSubscriptionFlowStep represents the step for deleting a balance subscription
	DeleteBalanceSubscriptionFlowStep FlowStep = "delete_balance_subscription"
	// ChooseBalanceSubscriptionToDeleteFlowStep represents the step for choosing balance subscription to delete
	ChooseBalanceSubscriptionToDeleteFlowStep FlowStep = "choose_balance_subscription_to_delete"
	// ConfirmDeleteBalanceSubscriptionFlowStep represents the step for confirming deletion of a balance subscription
	ConfirmDeleteBalanceSubscriptionFlowStep FlowStep = "confirm_delete_balance_subscription"
)
