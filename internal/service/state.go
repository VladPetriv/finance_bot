package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/VladPetriv/finance_bot/internal/model"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type stateService struct {
	logger *logger.Logger
	stores Stores
	apis   APIs
}

var _ StateService = (*stateService)(nil)

// StateOptions represents input options for creating new instance of state service.
type StateOptions struct {
	Logger *logger.Logger
	Stores Stores
	APIs   APIs
}

// NewState returns new instance of state service.
func NewState(opts *StateOptions) *stateService {
	return &stateService{
		logger: opts.Logger,
		stores: opts.Stores,
		apis:   opts.APIs,
	}
}

func (s stateService) HandleState(ctx context.Context, message Message) (*HandleStateOutput, error) {
	logger := s.logger.With().Str("name", "stateService.HandleState").Logger()
	logger.Debug().
		Any("message", message.GetText()).
		Any("sender", message.GetSenderName()).
		Any("chat_id", message.GetChatID()).
		Msg("handling message")

	user, err := s.stores.User.Get(ctx, GetUserFilter{
		Username:        message.GetSenderName(),
		PreloadSettings: true,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return nil, fmt.Errorf("get user from store: %w", err)
	}

	event := getEventFromMsg(user, message)
	logger.Debug().Any("event", event).Msg("got event based on bot message")

	state, err := s.stores.State.Get(ctx, GetStateFilter{
		UserID: message.GetSenderName(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get state from store")
		return nil, fmt.Errorf("get state from store: %w", err)
	}
	logger.Debug().Any("state", state).Msg("got state from store")

	if state == nil {
		return s.handleNewState(ctx, message, event, logger)
	}

	// Handle simple events that require immediate completion
	if s.isSimpleEvent(event) {
		return s.handleSimpleEvent(ctx, message, state, event)
	}

	// Handle unfinished flows
	if !state.IsFlowFinished() && isBotCommand(message.GetText()) && !state.IsCommandAllowedDuringFlow(message.GetText()) {
		return s.handleUnfinishedFlow(message, state)
	}

	// Handle ongoing flow or create new state
	return s.handleOngoingFlow(ctx, message, state, event)
}

func (s stateService) handleNewState(ctx context.Context, message Message, event model.Event, logger zerolog.Logger) (*HandleStateOutput, error) {
	newState := &model.State{
		ID:        uuid.NewString(),
		UserID:    message.GetSenderName(),
		Flow:      model.EventToFlow[event],
		Steps:     []model.FlowStep{model.StartFlowStep},
		Metedata:  make(map[string]any),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if s.isSimpleEvent(event) {
		newState.Steps = append(newState.Steps, model.EndFlowStep)
	}

	if isBotCommand(message.GetText()) {
		firstStep, ok := model.CommandToFistFlowStep[message.GetText()]
		if ok {
			newState.Steps = append(newState.Steps, firstStep)
		}
	}

	if event == model.CreateOperationsThroughOneTimeInputEvent {
		newState.Steps = append(newState.Steps, model.CreateOperationsThroughOneTimeInputFlowStep)
	}

	err := s.stores.State.Create(ctx, newState)
	if err != nil {
		logger.Error().Err(err).Msg("create state in store")
		return nil, fmt.Errorf("create state in store: %w", err)
	}

	logger.Info().Any("state", newState).Msg("created new state")
	return &HandleStateOutput{State: newState, Event: event}, nil
}

func (s stateService) isSimpleEvent(event model.Event) bool {
	return slices.Contains([]model.Event{
		model.CancelEvent,
		model.BackEvent,
		model.BalanceEvent,
		model.CategoryEvent,
		model.OperationEvent,
		model.BalanceSubscriptionEvent,
	}, event)
}

func (s stateService) handleSimpleEvent(ctx context.Context, message Message, state *model.State, event model.Event) (*HandleStateOutput, error) {
	logger := s.logger.With().Str("name", "stateService.handleSimpleEvent").Logger()

	err := s.stores.State.Delete(ctx, state.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete state from store")
		return nil, fmt.Errorf("delete state from store: %w", err)
	}

	completedState := &model.State{
		ID:        uuid.NewString(),
		UserID:    message.GetSenderName(),
		Flow:      model.EventToFlow[event],
		Steps:     []model.FlowStep{model.StartFlowStep, model.EndFlowStep},
		Metedata:  make(map[string]any),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// Store the base flow in metadata when canceling to preserve the user's context.
	// This allows returning to the previous keyboard layout instead of the default one.
	if event == model.CancelEvent {
		completedState.Metedata = map[string]any{
			baseFlowKey: model.GetBaseFlowFromCurrentFlow(state.Flow),
		}
	}

	err = s.stores.State.Create(ctx, completedState)
	if err != nil {
		logger.Error().Err(err).Msg("create state in store")
		return nil, fmt.Errorf("create state in store: %w", err)
	}

	return &HandleStateOutput{State: completedState, Event: event}, nil
}

func (s stateService) handleUnfinishedFlow(message Message, state *model.State) (*HandleStateOutput, error) {
	logger := s.logger.With().Str("name", "stateService.handleUnfinishedFlow").Logger()

	err := s.apis.Messenger.SendMessage(
		message.GetChatID(),
		fmt.Sprintf("You're previous flow(%s) is not finished. Please, finish it or cancel it before running new one.", state.GetFlowName()),
	)
	if err != nil {
		logger.Error().Err(err).Msg("send message to user")
		return nil, fmt.Errorf("send message to user: %w", err)
	}

	return nil, nil
}

func (s stateService) handleOngoingFlow(ctx context.Context, message Message, state *model.State, event model.Event) (*HandleStateOutput, error) {
	if event == model.UnknownEvent && !state.IsFlowFinished() {
		return &HandleStateOutput{
			State: state,
			Event: state.GetEvent(),
		}, nil
	}
	if event == model.CreateOperationsThroughOneTimeInputEvent && !state.IsFlowFinished() {
		return &HandleStateOutput{
			State: state,
			Event: state.GetEvent(),
		}, nil
	}

	if event != model.UnknownEvent && state.IsFlowFinished() {
		return s.createNewFlow(ctx, message, state, event)
	}

	return &HandleStateOutput{
		State: nil,
		Event: model.UnknownEvent,
	}, nil
}

func (s stateService) createNewFlow(ctx context.Context, message Message, state *model.State, event model.Event) (*HandleStateOutput, error) {
	logger := s.logger.With().Str("name", "stateService.createNewFlow").Logger()

	err := s.stores.State.Delete(ctx, state.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete state from store")
		return nil, fmt.Errorf("delete state from store: %w", err)
	}

	return s.handleNewState(ctx, message, event, logger)
}

func isBotCommand(value string) bool {
	return slices.Contains(model.AvailableCommands, value)
}

func (s stateService) DeleteState(ctx context.Context, message Message) error {
	logger := s.logger.With().Str("name", "stateService.DeleteState").Logger()

	state, err := s.stores.State.Get(ctx, GetStateFilter{
		UserID: message.GetSenderName(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get state from store")
		return fmt.Errorf("get state from store: %w", err)
	}
	if state == nil {
		logger.Info().Msg("state not found, no deletion needed")
		return nil
	}
	logger.Debug().Any("state", state).Msg("got state from store")

	if state.IsFlowFinished() {
		logger.Warn().Msg("deleting not finished state")
	}

	err = s.stores.State.Delete(ctx, state.ID)
	if err != nil {
		logger.Error().Err(err).Msg("delete state from store")
		return fmt.Errorf("delete state from store: %w", err)
	}

	return nil
}
