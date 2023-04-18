package service

import (
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
)

type eventService struct {
	botAPI bot.API
	logger *logger.Logger
}

var _ EventService = (*eventService)(nil)

// NewEvent returns new instance of event service.
func NewEvent(botAPI bot.API, logger *logger.Logger) *eventService {
	return &eventService{
		botAPI: botAPI,
		logger: logger,
	}
}

func (e eventService) Listen() error {
	return nil
}
