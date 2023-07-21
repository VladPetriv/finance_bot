package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/models"
	"github.com/VladPetriv/finance_bot/pkg/bot"
	"github.com/VladPetriv/finance_bot/pkg/logger"
	"github.com/VladPetriv/finance_bot/pkg/money"
	"github.com/google/uuid"
)

type handlerService struct {
	logger           *logger.Logger
	messageService   MessageService
	keyboardService  KeyboardService
	categoryService  CategoryService
	categoryStore    CategoryStore
	userService      UserService
	balanceStore     BalanceStore
	balanceService   BalanceService
	operationService OperationService
	operationStore   OperationStore
}

var _ HandlerService = (*handlerService)(nil)

// HandlerOptions represents input options for new instance of handler service.
type HandlerOptions struct {
	Logger           *logger.Logger
	MessageService   MessageService
	KeyboardService  KeyboardService
	CategoryService  CategoryService
	UserService      UserService
	BalanceStore     BalanceStore
	BalanceService   BalanceService
	OperationService OperationService
	OperationStore   OperationStore
	CategoryStore    CategoryStore
}

// NewHandler returns new instance of handler service.
func NewHandler(opts *HandlerOptions) *handlerService {
	return &handlerService{
		logger:           opts.Logger,
		messageService:   opts.MessageService,
		keyboardService:  opts.KeyboardService,
		categoryService:  opts.CategoryService,
		userService:      opts.UserService,
		balanceStore:     opts.BalanceStore,
		balanceService:   opts.BalanceService,
		operationService: opts.OperationService,
		operationStore:   opts.OperationStore,
		categoryStore:    opts.CategoryStore,
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

	user := &models.User{
		ID:       uuid.NewString(),
		Username: msg.Message.From.Username,
	}

	err = h.userService.CreateUser(ctx, user)
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

	err = h.balanceStore.Create(ctx, &models.Balance{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Amount: "0",
	})
	if err != nil {
		logger.Error().Err(err).Msg("create balance")
		return fmt.Errorf("create balance: %w", err)
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

	if IsBotCommand(msg.Message.Text) {
		err = h.messageService.SendMessage(&SendMessageOptions{
			ChatID: msg.Message.Chat.ID,
			Text:   "Enter category name!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		// TODO: What if user not found?
		logger.Error().Err(err).Msg("get user by username")
		return fmt.Errorf("get user by username: %w", err)
	}

	err = h.categoryService.CreateCategory(ctx, &models.Category{
		ID:     uuid.NewString(),
		UserID: user.ID,
		Title:  msg.Message.Text,
	})
	if err != nil {
		if errors.Is(err, ErrCategoryAlreadyExists) {
			err = h.messageService.SendMessage(&SendMessageOptions{
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

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		// TODO: What if user not found?
		logger.Error().Err(err).Msg("get user by username")
		return fmt.Errorf("get user by username: %w", err)
	}

	// TODO: Pass context into all handler functions.
	categories, err := h.categoryService.ListCategories(ctx, user.ID)
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

func (h handlerService) HandleEventUpdateBalance(ctx context.Context, eventName event, messageData []byte) error {
	logger := h.logger

	var msg HandleEventUpdateBalance

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event update balance message")
		return fmt.Errorf("unmarshal handle event update balance message: %w", err)
	}

	isBotCommand := IsBotCommand(msg.Message.Text)

	if isBotCommand && eventName == updateBalanceEvent {
		err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.Message.Chat.ID,
			Message: "Choose what you want to update in you balance:",
			Type:    keyboardTypeRow,
			Rows: []bot.KeyboardRow{
				{Buttons: []string{botUpdateBalanceAmountCommand, botUpdateBalanceCurrencyCommand}},
				{Buttons: []string{botBackCommand}},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user by username")
		return fmt.Errorf("get user by username: %w", err)
	}

	balance, err := h.balanceStore.Get(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance by user id")
		return fmt.Errorf("get balance by user id")
	}

	if eventName == updateBalanceAmountEvent {
		err = h.handleUpdateBalanceAmountEvent(ctx, updateBalanceAmountOptions{
			balance:      balance,
			chatID:       msg.Message.Chat.ID,
			amount:       msg.Message.Text,
			isBotCommand: isBotCommand,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle update balance amount event")
			return fmt.Errorf("handle update balance amount event: %w", err)
		}
	}
	if eventName == updateBalanceCurrencyEvent {
		err = h.handleUpdateBalanceCurrencyEvent(ctx, updateBalanceCurrencyOptions{
			balance:      balance,
			chatID:       msg.Message.Chat.ID,
			currency:     msg.Message.Text,
			isBotCommand: isBotCommand,
		})
		if err != nil {
			logger.Error().Err(err).Msg("handle update balance currency event")
			return fmt.Errorf("handle update balance currency event: %w", err)
		}
	}

	logger.Info().Msg("balance successfully updated")
	return nil
}

type updateBalanceAmountOptions struct {
	balance      *models.Balance
	chatID       int64
	amount       string
	isBotCommand bool
}

func (h handlerService) handleUpdateBalanceAmountEvent(ctx context.Context, opts updateBalanceAmountOptions) error {
	logger := h.logger

	if opts.isBotCommand {
		err := h.messageService.SendMessage(&SendMessageOptions{
			ChatID: opts.chatID,
			Text:   "Enter balance amount:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}

	price, err := money.NewFromString(opts.amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert string to money type")

		err = h.messageService.SendMessage(&SendMessageOptions{
			ChatID: opts.chatID,
			Text:   "Please enter amount in the right format!\nExamples: 1000.12, 10.12, 35",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}

	opts.balance.Amount = price.String()

	err = h.balanceStore.Update(ctx, opts.balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance")
		return fmt.Errorf("update balance: %w", err)
	}

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: opts.chatID,
		Text:   "Balance values successfully updated!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("balance amount successfully updated")
	return nil
}

type updateBalanceCurrencyOptions struct {
	balance      *models.Balance
	chatID       int64
	currency     string
	isBotCommand bool
}

func (h handlerService) handleUpdateBalanceCurrencyEvent(ctx context.Context, opts updateBalanceCurrencyOptions) error {
	logger := h.logger

	if opts.isBotCommand {
		err := h.messageService.SendMessage(&SendMessageOptions{
			ChatID: opts.chatID,
			Text:   "Enter balance currency:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	}

	opts.balance.Currency = opts.currency

	err := h.balanceStore.Update(ctx, opts.balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance")
		return fmt.Errorf("update balance: %w", err)
	}

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: opts.chatID,
		Text:   "Balance currency successfully updated!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("balance amount successfully updated")
	return nil
}

func (h handlerService) HandleEventGetBalance(ctx context.Context, messageData []byte) error {
	logger := h.logger

	var msg HandleEventGetBalance
	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event get balance message")
		return fmt.Errorf("unmarshal handle event get balance message: %w", err)
	}

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user by usernmae")
		return fmt.Errorf("get user by usernmae: %w", err)
	}
	logger.Debug().Interface("user", user).Msg("got user by username")

	balanceInfo, err := h.balanceService.GetBalanceInfo(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balcne info by user id")
		return fmt.Errorf("get balance info by user id")
	}
	logger.Debug().Interface("balanceInfo", balanceInfo).Msg("got balance info by user id")

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text: fmt.Sprintf(
			"Hello, @%s!\nYour current balance is: %v%s!",
			user.Username, balanceInfo.Amount, balanceInfo.Currency,
		),
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("successfully handled get balance event")
	return nil
}

func (h handlerService) HandleEventOperationCreate(ctx context.Context, eventName event, messageData []byte) error {
	logger := h.logger

	var msg HandleEventOperationCreate

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event update balance message")
		return fmt.Errorf("unmarshal handle event update balance message: %w", err)
	}

	isBotCommand := IsBotCommand(msg.Message.Text) || IsBotCommand(msg.CallbackQuery.Data)
	if isBotCommand && eventName == createOperationEvent {
		err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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

		return nil
	}

	user, err := h.userService.GetUserByUsername(ctx, msg.GetUsername())
	if err != nil {
		logger.Error().Err(err).Msg("get user by username")
		return fmt.Errorf("get user by username: %w", err)
	}

	balance, err := h.balanceStore.Get(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance from storage")
		return fmt.Errorf("get balance from storage: %w", err)
	}

	if msg.CallbackQuery.Data != "" && (eventName == createIncomingOperationEvent || eventName == createSpendingOperationEvent) {
		categories, err := h.categoryService.ListCategories(ctx, user.ID)
		if err != nil {
			if errors.Is(err, ErrCategoriesNotFound) {
				err = h.messageService.SendMessage(&SendMessageOptions{
					ChatID: msg.GetChatID(),
					Text:   "Please create a category before creating operation",
				})
				if err != nil {
					logger.Error().Err(err).Msg("send message")
					return fmt.Errorf("send message: %w", err)
				}
				return nil
			}

			logger.Error().Err(err).Msg("list categories")
			return fmt.Errorf("list categories: %w", err)
		}

		keyboardRows := []bot.KeyboardRow{
			{Buttons: []string{}},
			{Buttons: []string{botBackCommand}},
		}

		for _, c := range categories {
			keyboardRows[0].Buttons = append(keyboardRows[0].Buttons, c.Title)
		}

		err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.GetChatID(),
			Message: "Choose category",
			Type:    keyboardTypeRow,
			Rows:    keyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		return nil
	}

	if msg.Message.Text != "" {
		category, err := h.categoryStore.GetByTitle(ctx, msg.Message.Text)
		if err != nil {
			logger.Error().Err(err).Msg("get category by title from storage")
			return fmt.Errorf("get category by title from storage: %w", err)
		}
		if category == nil {
			err = h.messageService.SendMessage(&SendMessageOptions{
				ChatID: msg.GetChatID(),
				Text:   "Category not found please try again!",
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}
			return nil
		}

		var operationType models.OperationType
		if eventName == createIncomingOperationEvent {
			operationType = models.OperationTypeIncoming
		}
		if eventName == createSpendingOperationEvent {
			operationType = models.OperationTypeSpending
		}
		err = h.operationStore.Create(ctx, &models.Operation{
			ID:         uuid.NewString(),
			BalanceID:  balance.ID,
			CategoryID: category.ID,
			Type:       operationType,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create operation in storage")
			return fmt.Errorf("create operation in storage: %w", err)
		}

		err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}
	}

	logger.Info().Msg("operation successfully created")
	return nil
}

func (h handlerService) HandleEventUpdateOperationAmount(ctx context.Context, messageData []byte) error {
	logger := h.logger

	var msg HandleEventUpdateOperationAmount
	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event get balance message")
		return fmt.Errorf("unmarshal handle event get balance message: %w", err)
	}

	if IsBotCommand(msg.Message.Text) || msg.Message.Text == botUpdateOperationAmountCommand {
		err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.Message.Chat.ID,
			Message: "Enter operation amount!",
			Type:    keyboardTypeRow,
			Rows:    defaultKeyboardRows,
		})
		if err != nil {
			logger.Error().Err(err).Msg("create row keyboard")
			return fmt.Errorf("create row keyboard: %w", err)
		}

		return nil
	}

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user by usernmae")
		return fmt.Errorf("get user by usernmae: %w", err)
	}

	balanceInfo, err := h.balanceService.GetBalanceInfo(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balcne info by user id")
		return fmt.Errorf("get balance info by user id")
	}

	operations, err := h.operationStore.GetAll(ctx, balanceInfo.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get operations from storage")
		return fmt.Errorf("get operations from storage: %w", err)
	}
	if len(operations) == 0 {
		err = h.messageService.SendMessage(&SendMessageOptions{
			ChatID: msg.Message.Chat.ID,
			Text:   "Operation not found please try again!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}
	}

	if operations[len(operations)-1].Amount == "" {
		operations[len(operations)-1].Amount = msg.Message.Text
		err := h.operationService.CreateOperation(ctx, CreateOperationOptions{
			UserID:    user.ID,
			Operation: &operations[len(operations)-1],
		})
		if err != nil {
			logger.Error().Err(err).Msg("create operation")

			err = h.messageService.SendMessage(&SendMessageOptions{
				ChatID: msg.Message.Chat.ID,
				Text:   "Can't create operation!",
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}
		}
	}

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   "Operation successfully created",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (h handlerService) HandleEventBack(ctx context.Context, messageData []byte) error {
	logger := h.logger

	var msg HandleEventUnknownMessage

	err := json.Unmarshal(messageData, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("unmarshal handle event unknown message")
		return fmt.Errorf("unmarshal event unknown message: %w", err)
	}
	logger.Debug().Interface("msg", msg).Msg("unmarshalled handle event unknown message")

	err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.Message.Chat.ID,
		Message: "Please choose command to execute:",
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard")
		return fmt.Errorf("create keyboard: %w", err)
	}

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
