package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
)

type handlerService struct {
	logger          *logger.Logger
	messageService  MessageService
	keyboardService KeyboardService
	categoryService CategoryService
	userService     UserService
}

var _ HandlerService = (*handlerService)(nil)

// HandlerOptions represents input options for new instance of handler service.
type HandlerOptions struct {
	Logger          *logger.Logger
	MessageService  MessageService
	KeyboardService KeyboardService
	CategoryService CategoryService
	UserService     UserService
}

// NewHandler returns new instance of handler service.
func NewHandler(opts *HandlerOptions) *handlerService {
	return &handlerService{
		logger:          opts.Logger,
		messageService:  opts.MessageService,
		keyboardService: opts.KeyboardService,
		categoryService: opts.CategoryService,
		userService:     opts.UserService,
	}
}

func (h handlerService) HandleEventStart(ctx context.Context, messageData []byte) error {
	logger := h.logger

	var msg HandleEventStartMessage

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event start message")
		return fmt.Errorf("unmarshal event start message: %w", err)
	}
	logger.Debug().Interface("msg", msg).Msg("unmarshalled handle event start message")

	welcomeMessage := fmt.Sprintf("Hello, @%s!\nWelcome to @FinanceTracking_bot!", msg.Message.From.Username)

	err = h.userService.CreateUser(ctx, &models.User{
		ID:       uuid.NewString(),
		Username: msg.Message.From.Username,
	})
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			welcomeMessage = fmt.Sprintf("Happy to see you again @%s!", msg.Message.From.Username)

			err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
				ChatID:  msg.Message.Chat.ID,
				Message: welcomeMessage,
				Type:    keyboardTypeRow,
				Rows:    defaultKeyboardRows,
			})
			if err != nil {
				logger.Error().Err(err).Msg("create keyboard with message")
				return fmt.Errorf("create keyboard: %w", err)
			}

			logger.Info().Msg("got already known user")
			return nil
		}
		logger.Error().Err(err).Msg("create user")
		return fmt.Errorf("create user: %w", err)
	}

	err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.Message.Chat.ID,
		Message: welcomeMessage,
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard with message")
		return fmt.Errorf("create keyboard: %w", err)
	}

	return nil
}

func (h handlerService) HandleEventCategoryCreate(ctx context.Context, messageData []byte) error {
	logger := h.logger

	var msg HandleEventCategoryCreate

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event category create message")
		return fmt.Errorf("unmarshal event category create message: %w", err)
	}
	logger.Debug().Interface("msg", msg).Msg("unmarshalled handle event category create message")

	if len(msg.Message.Entities) != 0 && msg.Message.Entities[0].IsBotCommand() {
		err := h.messageService.SendMessage(&SendMessageOptions{
			ChatID: msg.Message.Chat.ID,
			Text:   "Enter category name!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}

	err = h.categoryService.CreateCategory(ctx, &models.Category{
		ID:    uuid.NewString(),
		Title: msg.Message.Text,
	})
	if err != nil {
		if errors.Is(err, ErrCategoryAlreadyExists) {
			err := h.messageService.SendMessage(&SendMessageOptions{
				ChatID: msg.Message.Chat.ID,
				Text:   fmt.Sprintf("Category with name '%s' already exists!", msg.Message.Text),
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}
		}

		return fmt.Errorf("send message: %w", err)
	}

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Category successfully created!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("successfully handle create category event")
	return nil
}

func (h handlerService) HandleEventListCategories(ctx context.Context, messageData []byte) error {
	logger := h.logger

	var msg HandleEventListCategories

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event list categories message")
		return fmt.Errorf("unmarshal event list categories message: %w", err)
	}
	logger.Debug().Interface("msg", msg).Msg("unmarshalled handle event list categories message")

	// TODO: Pass context into all handler functions.
	categories, err := h.categoryService.ListCategories(ctx)
	if err != nil {
		if errors.Is(err, ErrCategoriesNotFound) {
			err = h.messageService.SendMessage(&SendMessageOptions{
				ChatID: msg.Message.Chat.ID,
				Text:   "Categories not found!",
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}

			return nil
		}
		logger.Error().Err(err).Msg("get list of categories")
		return fmt.Errorf("get list of categories: %w", err)
	}

	outputMessage := "Categories: \n"

	for i, c := range categories {
		i++
		outputMessage += fmt.Sprintf("%v. %s\n", i, c.Title)
	}

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   outputMessage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("successfully handle list categories event")
	return nil
}

func (h handlerService) HandleEventUnknown(messageData []byte) error {
	logger := h.logger

	var msg HandleEventUnknownMessage

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event unknown message")
		return fmt.Errorf("unmarshal event unknown message: %w", err)
	}
	logger.Debug().Interface("msg", msg).Msg("unmarshalled handle event unknown message")

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Didn't understand you!\nCould you please check available commands!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
