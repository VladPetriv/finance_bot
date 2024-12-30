package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
)

type stateService struct {
	logger *logger.Logger
	stores Stores
}

var _ StateService = (*stateService)(nil)

// StateOptions represents input options for creating new instance of state service.
type StateOptions struct {
	Logger *logger.Logger
	Stores Stores
}

// NewState returns new instance of state service.
func NewState(opts *StateOptions) *stateService {
	return &stateService{
		logger: opts.Logger,
		stores: opts.Stores,
	}
}

func (s stateService) HandleState(ctx context.Context, message botMessage) (*HandleStateOutput, error) {
	logger := s.logger.With().Str("name", "stateService.HandleState").Logger()
	logger.Debug().Any("message", message).Msg("got args")

	state, err := s.stores.State.Get(ctx, GetStateFilter{
		UserID: message.GetUsername(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get state from store")
		return nil, fmt.Errorf("get state from store: %w", err)
	}

	event := getEventFromMsg(&message)
	logger.Debug().Any("event", event).Msg("got event based on bot message")

	if state == nil {
		flow := models.EventToFlow[event]

		stateForCreate := &models.State{
			ID:        uuid.NewString(),
			UserID:    message.GetUsername(),
			Flow:      flow,
			Steps:     []models.FlowStep{models.StartFlowStep},
			Metedata:  make(map[string]any),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if isBotCommand(message.Message.Text) {
			firstFlowStep := models.CommadToFistFlowStep[message.Message.Text]

			// NOTE: We handle here only flows that require more than two step.
			if firstFlowStep != "" {
				stateForCreate.Steps = append(stateForCreate.Steps, firstFlowStep)
			}
		}

		err := s.stores.State.Create(ctx, stateForCreate)
		if err != nil {
			logger.Error().Err(err).Msg("create state in store")
			return nil, fmt.Errorf("create state in store: %w", err)
		}

		logger.Info().Any("state", stateForCreate).Msg("created new state")
		return &HandleStateOutput{
			State: stateForCreate,
			Event: event,
		}, nil
	}
	logger.Debug().Any("state", state).Msg("got state from store")

	// If we're not able to define the event based on message text and flow is not finished yet
	// we should return the same state, since current flow is not finished and myabe we process other steps.
	if event == models.UnknownEvent && !state.IsFlowFinished() {
		return &HandleStateOutput{
			State: state,
			Event: state.GetEvent(),
		}, nil
	}

	// Received from database state is with finished flow and event that was received from message is not uknown.
	// We should delete current stateand create new one with initial flow step.
	if event != models.UnknownEvent && state.IsFlowFinished() {
		err := s.stores.State.Delete(ctx, state.ID)
		if err != nil {
			logger.Error().Err(err).Msg("delete state from store")
			return nil, fmt.Errorf("delete state from store: %w", err)
		}

		stateForCreate := &models.State{
			ID:        uuid.NewString(),
			UserID:    message.GetUsername(),
			Flow:      models.EventToFlow[event],
			Steps:     []models.FlowStep{models.StartFlowStep},
			Metedata:  make(map[string]any),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if isBotCommand(message.Message.Text) {
			firstFlowStep := models.CommadToFistFlowStep[message.Message.Text]

			// NOTE: We handle here only flows that require more than two step.
			if firstFlowStep != "" {
				stateForCreate.Steps = append(stateForCreate.Steps, firstFlowStep)
			}
		}

		err = s.stores.State.Create(ctx, stateForCreate)
		if err != nil {
			logger.Error().Err(err).Msg("create state in store")
			return nil, fmt.Errorf("create state in store: %w", err)
		}

		logger.Info().Any("state", stateForCreate).Msg("created new state")
		return &HandleStateOutput{
			State: stateForCreate,
			Event: event,
		}, nil
	}

	return &HandleStateOutput{
		State: nil,
		Event: models.UnknownEvent,
	}, nil
}

func isBotCommand(command string) bool {
	return strings.Contains(strings.Join(models.AvailableCommands, " "), command)
}
