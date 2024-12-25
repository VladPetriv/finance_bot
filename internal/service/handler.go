package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/google/uuid"
)

type handlerService struct {
	logger   *logger.Logger
	services Services
	stores   Stores
}

var _ HandlerService = (*handlerService)(nil)

// HandlerOptions represents input options for new instance of handler service.
type HandlerOptions struct {
	Logger   *logger.Logger
	Services Services
	Stores   Stores
}

// NewHandler returns new instance of handler service.
func NewHandler(opts *HandlerOptions) *handlerService {
	return &handlerService{
		logger:   opts.Logger,
		services: opts.Services,
		stores:   opts.Stores,
	}
}

func (h handlerService) HandleEventStart(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	var nextStep models.FlowStep
	defer func() {
		state := ctx.Value(contextFieldNameState).(*models.State)
		state.Steps = append(state.Steps, nextStep)
		updatedState, err := h.stores.State.Update(ctx, state)
		if err != nil {
			logger.Error().Err(err).Msg("update state in store")
			return
		}
		logger.Debug().Interface("updatedState", updatedState).Msg("updated state in store")
	}()

	username := msg.GetUsername()
	chatID := msg.GetChatID()
	logger.Debug().
		Str("username", username).
		Int64("chatID", chatID).
		Msg("got username and chat id from incoming message")

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

	// Handle case when user already exists
	if user != nil {
		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  chatID,
			Message: fmt.Sprintf("Happy to see you again @%s!", username),
			Type:    keyboardTypeRow,
			Rows:    defaultKeyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard with welcome message")
			return fmt.Errorf("create keyboard with welcome message: %w", err)
		}

		nextStep = models.EndFlowStep

		logger.Info().Msg("user already exists")
		return nil
	}

	err = h.stores.User.Create(ctx, &models.User{
		ID:       uuid.NewString(),
		Username: username,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create user in store")
		return fmt.Errorf("create user in store: %w", err)
	}

	welcomeMessage := fmt.Sprintf("Hello, @%s!\nWelcome to @FinanceTracking_bot!", username)
	enterBalanceNameMessage := "Please enter the name of your initial balance!:"

	messagesToSend := []string{welcomeMessage, enterBalanceNameMessage}
	for _, message := range messagesToSend {
		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: chatID,
			Text:   message,
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}
	}

	nextStep = models.CreateInitialBalanceFlowStep

	logger.Info().Msg("handled event start")
	return nil
}

func (h handlerService) HandleEventOperationCreate(ctx context.Context, eventName event, msg botMessage) error {
	logger := h.logger

	isBotCommand := IsBotCommand(msg.Message.Text) || IsBotCommand(msg.CallbackQuery.Data)

	if isBotCommand && eventName == createOperationEvent {
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose operation type",
			Type:    keyboardTypeInline,
			Rows: []bot.KeyboardRow{
				{Buttons: []string{botCreateIncomingOperationCommand, botCreateSpendingOperationCommand}},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create inline keyboard")
			return fmt.Errorf("create inline keyboard: %w", err)
		}

		logger.Info().Msg("handle command input")
		return nil
	}

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: msg.GetUsername(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user by username")
		return fmt.Errorf("get user by username: %w", err)
	}

	balance, err := h.stores.Balance.Get(ctx, GetBalanceFilter{
		UserID: user.ID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
	}

	if msg.CallbackQuery.Data != "" && (eventName == createIncomingOperationEvent || eventName == createSpendingOperationEvent) {
		// categories, err := h.services.Category.ListCategories(ctx, user.ID)
		// if err != nil {
		// 	if errors.Is(err, ErrCategoriesNotFound) {
		// 		err = h.services.Message.SendMessage(&SendMessageOptions{
		// 			ChatID: msg.GetChatID(),
		// 			Text:   "Please create a category before creating operation",
		// 		})
		// 		if err != nil {
		// 			logger.Error().Err(err).Msg("send message")
		// 			return fmt.Errorf("send message: %w", err)
		// 		}

		// 		logger.Info().Msg("categories not found")
		// 		return nil
		// 	}

		// 	logger.Error().Err(err).Msg("get list of categories from store")
		// 	return fmt.Errorf("get list of categories from store: %w", err)
		// }

		keyboardRows := []bot.KeyboardRow{
			{Buttons: []string{}},
			{Buttons: []string{botBackCommand}},
		}

		// for _, c := range categories {
		// 	keyboardRows[0].Buttons = append(keyboardRows[0].Buttons, c.Title)
		// }

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose category",
			Type:    keyboardTypeRow,
			Rows:    keyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		logger.Info().Msg("handled second command input")
		return nil
	}

	if msg.Message.Text != "" {
		category, err := h.stores.Category.Get(ctx, GetCategoryFilter{
			Title: msg.Message.Text,
		})
		if err != nil {
			logger.Error().Err(err).Msg("get category store")
			return fmt.Errorf("get category store: %w", err)
		}
		if category == nil {
			err = h.services.Message.SendMessage(&SendMessageOptions{
				ChatID: msg.GetChatID(),
				Text:   "Category not found please try again!",
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}

			logger.Info().Msg("category not found")
			return nil
		}

		var operationType models.OperationType
		if eventName == createIncomingOperationEvent {
			operationType = models.OperationTypeIncoming
		}
		if eventName == createSpendingOperationEvent {
			operationType = models.OperationTypeSpending
		}

		err = h.stores.Operation.Create(ctx, &models.Operation{
			ID:         uuid.NewString(),
			BalanceID:  balance.ID,
			CategoryID: category.ID,
			Type:       operationType,
			CreatedAt:  time.Now(),
		})
		if err != nil {
			logger.Error().Err(err).Msg("create operation in store")
			return fmt.Errorf("create operation in store: %w", err)
		}

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Please click on the button bellow for entering operation amount!",
			Type:    keyboardTypeRow,
			Rows: []bot.KeyboardRow{
				{
					Buttons: []string{botUpdateOperationAmountCommand},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}
	}

	logger.Info().Msg("handle event operation create")
	return nil
}

func (h handlerService) HandleEventUpdateOperationAmount(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	if IsBotCommand(msg.Message.Text) || msg.Message.Text == botUpdateOperationAmountCommand {
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.Message.Chat.ID,
			Message: "Enter operation amount!",
			Type:    keyboardTypeRow,
			Rows:    defaultKeyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		logger.Info().Msg("handle command input")
		return nil
	}

	user, err := h.stores.User.Get(ctx, GetUserFilter{
		Username: msg.GetUsername(),
	})
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

	// TODO: Replace with store call
	var balanceInfo *models.Balance

	operations, err := h.stores.Operation.GetAll(ctx, balanceInfo.ID, GetAllOperationsFilter{})
	if err != nil {
		logger.Error().Err(err).Msg("get all operations from store")
		return fmt.Errorf("get all operations from store: %w", err)
	}
	if len(operations) == 0 {
		err = h.services.Message.SendMessage(&SendMessageOptions{
			ChatID: msg.Message.Chat.ID,
			Text:   "Operation not found please try again!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		logger.Info().Msg("operations not found")
		return nil
	}

	if operations[len(operations)-1].Amount == "" {
		operations[len(operations)-1].Amount = msg.Message.Text
		err = h.services.Operation.CreateOperation(ctx, CreateOperationOptions{
			UserID:    user.ID,
			Operation: &operations[len(operations)-1],
		})
		if err != nil {
			if errors.Is(err, ErrInvalidAmountFormat) {
				err = h.services.Message.SendMessage(&SendMessageOptions{
					ChatID: msg.GetChatID(),
					Text:   "Please enter amount in the right format!\nExamples: 1000.12, 10.12, 35",
				})
				if err != nil {
					logger.Error().Err(err).Msg("send message")
					return fmt.Errorf("send message: %w", err)
				}

				return nil
			}

			logger.Error().Err(err).Msg("create operation in store")
			return fmt.Errorf("create operation in store: %w", err)
		}
	}

	err = h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Operation successfully created",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled event update operation amount")
	return nil
}

func (h handlerService) HandleEventGetOperationsHistory(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	if IsBotCommand(msg.Message.Text) {
		err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Please select a period for operation history!",
			Type:    keyboardTypeRow,
			Rows: []bot.KeyboardRow{
				{
					Buttons: []string{
						string(models.CreationPeriodDay),
						string(models.CreationPeriodWeek),
						string(models.CreationPeriodMonth),
						string(models.CreationPeriodYear),
					},
				},
				{
					Buttons: []string{botBackCommand},
				},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		return nil
	}

	// user, err := h.stores.User.Get(ctx, GetUserFilter{
	// 	Username: msg.GetUsername(),
	// })
	// if err != nil {
	// 	logger.Error().Err(err).Msg("get user from store")
	// 	return fmt.Errorf("get user from store: %w", err)
	// }

	// TODO: Replace with store call
	var balanceInfo *models.Balance

	creationPeriod := models.GetCreationPeriodFromText(msg.Message.Text)
	if creationPeriod == nil {
		logger.Error().Msgf("message text is not creation period, text: %s", msg.Message.Text)
		return fmt.Errorf("message text is not creation period")
	}

	operations, err := h.stores.Operation.GetAll(ctx, balanceInfo.ID, GetAllOperationsFilter{
		CreationPeriod: creationPeriod,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get all operations from store")
		return fmt.Errorf("get all operations from store: %w", err)
	}
	if operations == nil {
		logger.Info().Msg("operations not found")

		err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Type:    keyboardTypeRow,
			Rows:    defaultKeyboardRows,
			Message: "Operations during that period of time not found!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}

	resultMessage := fmt.Sprintf("Balance Amount: %v%s\nPeriod: %v\n", balanceInfo.Amount, balanceInfo.Currency, *creationPeriod)

	for _, o := range operations {
		resultMessage += fmt.Sprintf(
			"\nOperation: %s\nCategory: %s\nAmount: %v%s\nCreation date: %v\n- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -",
			o.Type, o.CategoryID, o.Amount, balanceInfo.Currency, o.CreatedAt.Format(time.ANSIC),
		)
	}

	err = h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.GetChatID(),
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
		Message: resultMessage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled event get operations history")
	return nil
}

func (h handlerService) HandleEventBack(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	err := h.services.Keyboard.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.Message.Chat.ID,
		Message: "Please choose command to execute:",
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create row keyboard")
		return fmt.Errorf("create row keyboard: %w", err)
	}

	logger.Info().Msg("handled event back")
	return nil
}

func (h handlerService) HandleEventUnknown(msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	err := h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Didn't understand you!\nCould you please check available commands!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled event back")
	return nil
}

func (h handlerService) HandleError(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	err := h.services.Message.SendMessage(&SendMessageOptions{
		ChatID: msg.GetChatID(),
		Text:   "Something went wrong!\nPlease try again later!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled error")
	return nil
}
