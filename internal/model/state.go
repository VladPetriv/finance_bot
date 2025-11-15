package model

import (
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

	Metadata Metadata `db:"metadata"`

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
	case StartFlow:
		switch s.GetCurrentStep() {
		case EnterBalanceCurrencyFlowStep:
			return slices.Contains(
				[]string{BotPreviousCommand, BotNextCommand},
				command,
			)
		}

		return false

	case UpdateUserSettingsFlow:
		if s.GetCurrentStep() == ChooseUpdateUserSettingsOptionFlowStep {
			return slices.Contains(
				[]string{BotUpdateUserAIParserCommand, BotUpdateUserSubscriptionNotificationsCommand},
				command,
			)
		}

		if s.GetCurrentStep() == UpdateAIParserEnabledUserSettingFlowStep || s.GetCurrentStep() == UpdateSubscriptionNotificationUserSettingFlowStep {
			return slices.Contains(
				[]string{BotEnableCommand, BotDisableCommand},
				command,
			)
		}

		return false

	case CreateOperationFlow:
		if s.GetCurrentStep() == ProcessOperationTypeFlowStep {
			return slices.Contains(
				[]string{BotCreateIncomingOperationCommand, BotCreateSpendingOperationCommand, BotCreateTransferOperationCommand},
				command,
			)
		}

		return false

	case GetOperationsHistoryFlow:
		if s.GetCurrentStep() == ChooseTimePeriodForOperationsHistoryFlowStep {
			return slices.Contains(
				[]string{BotPreviousCommand, BotNextCommand},
				command,
			)
		}

		return false

	case DeleteOperationFlow:
		if s.GetCurrentStep() == ChooseOperationToDeleteFlowStep {
			return slices.Contains(
				[]string{BotPreviousCommand, BotNextCommand},
				command,
			)
		}

		return false

	case UpdateOperationFlow:
		switch s.GetCurrentStep() {
		case ChooseOperationToUpdateFlowStep:
			return slices.Contains(
				[]string{BotPreviousCommand, BotNextCommand},
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
		case EnterBalanceCurrencyFlowStep:
			return slices.Contains(
				[]string{BotNextCommand, BotPreviousCommand},
				command,
			)
		case ChooseUpdateBalanceOptionFlowStep:
			return slices.Contains(
				[]string{
					BotUpdateBalanceNameCommand, BotUpdateBalanceAmountCommand, BotUpdateBalanceCurrencyCommand,
				},
				command,
			)
		}

		return false

	case CreateBalanceFlow:
		switch s.GetCurrentStep() {
		case EnterBalanceCurrencyFlowStep:
			return slices.Contains(
				[]string{BotNextCommand, BotPreviousCommand},
				command,
			)
		}

		return false

	case ListBalanceSubscriptionFlow:
		switch s.GetCurrentStep() {
		case ChooseBalanceFlowStep:
			return slices.Contains(
				[]string{BotNextCommand, BotPreviousCommand},
				command,
			)
		}

		return false

	case UpdateBalanceSubscriptionFlow:
		switch s.GetCurrentStep() {
		case ChooseBalanceSubscriptionToUpdateFlowStep:
			return slices.Contains(
				[]string{BotPreviousCommand, BotNextCommand},
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

	case DeleteBalanceSubscriptionFlow:
		switch s.GetCurrentStep() {
		case ChooseBalanceSubscriptionToDeleteFlowStep:
			return slices.Contains(
				[]string{BotPreviousCommand, BotNextCommand},
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

	if s.Flow == BackFlow && len(s.Steps) == 1 {
		return BackEvent
	}

	switch s.Steps[indexOfInitialFlowStep] {
	// User settings
	case GetUserSettingsFlowStep:
		return GetUserSettingsEvent
	case UpdateUserSettingsFlowStep:
		return UpdateUserSettingsEvent

	// Balance
	case CreateInitialBalanceFlowStep, CreateBalanceFlowStep:
		return CreateBalanceEvent
	case UpdateBalanceFlowStep:
		return UpdateBalanceEvent
	case GetBalanceFlowStep:
		return GetBalanceEvent
	case DeleteBalanceFlowStep:
		return DeleteBalanceEvent

	// Category
	case CreateCategoryFlowStep:
		return CreateCategoryEvent
	case ListCategoriesFlowStep:
		return ListCategoriesEvent
	case UpdateCategoryFlowStep:
		return UpdateCategoryEvent
	case DeleteCategoryFlowStep:
		return DeleteCategoryEvent

	// Operation
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

	// Balance Subscription
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
