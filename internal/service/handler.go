package service

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (h handlerService) HandleEventStart(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	welcomeMessage := fmt.Sprintf("Hello, @%s!\nWelcome to @FinanceTracking_bot!", msg.Message.From.Username)

	userID := uuid.NewString()

	err := h.userService.CreateUser(ctx, &models.User{
		ID:       userID,
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
				logger.Error().Err(err).Msg("create keyboard with welcome message")
				return fmt.Errorf("create keyboard with welcome message: %w", err)
			}

			logger.Info().Msg("user already exists")
			return nil
		}

		logger.Error().Err(err).Msg("create user in store")
		return fmt.Errorf("create user in store: %w", err)
	}

	err = h.balanceStore.Create(ctx, &models.Balance{
		ID:     uuid.NewString(),
		UserID: userID,
		Amount: "0",
	})
	if err != nil {
		logger.Error().Err(err).Msg("create balance in store")
		return fmt.Errorf("create balance in store: %w", err)
	}

	err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
		ChatID:  msg.Message.Chat.ID,
		Message: welcomeMessage,
		Type:    keyboardTypeRow,
		Rows:    defaultKeyboardRows,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create keyboard with welcome message")
		return fmt.Errorf("create keyboard with welcome message: %w", err)
	}

	logger.Info().Msg("handled event start")
	return nil
}

func (h handlerService) HandleEventCategoryCreate(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	if IsBotCommand(msg.Message.Text) {
		err := h.messageService.SendMessage(&SendMessageOptions{
			ChatID: msg.Message.Chat.ID,
			Text:   "Enter category name!",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		logger.Info().Msg("handled command input")
		return nil
	}

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
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

			logger.Info().Msg("category already exists")
			return nil
		}

		logger.Error().Err(err).Msg("send message")
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

	logger.Info().Msg("handled create category event")
	return nil
}

func (h handlerService) HandleEventListCategories(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

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

			logger.Info().Msg("categories not found")
			return nil
		}
		logger.Error().Err(err).Msg("get list of categories from store")
		return fmt.Errorf("get list of categories from store: %w", err)
	}

	outputMessage := "Categories: \n"

	for i, c := range categories {
		i++
		outputMessage += fmt.Sprintf("%v. %s\n", i, c.Title)
	}
	logger.Debug().Interface("outputMessage", outputMessage).Msg("built output message")

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: msg.Message.Chat.ID,
		Text:   outputMessage,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled list categories event")
	return nil
}

func (h handlerService) HandleEventUpdateBalance(ctx context.Context, eventName event, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	isBotCommand := IsBotCommand(msg.Message.Text)

	if isBotCommand && eventName == updateBalanceEvent {
		err := h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
			ChatID:  msg.Message.Chat.ID,
			Message: "Choose what you want to update in you balance:",
			Type:    keyboardTypeRow,
			Rows: []bot.KeyboardRow{
				{Buttons: []string{botUpdateBalanceAmountCommand, botUpdateBalanceCurrencyCommand}},
				{Buttons: []string{botBackCommand}},
			},
		})
		if err != nil {
			logger.Error().Err(err).Msg("create keyboard")
			return fmt.Errorf("create keyboard: %w", err)
		}

		logger.Info().Msg("handled command input")
		return nil
	}

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

	balance, err := h.balanceStore.Get(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
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

		logger.Info().Msg("handled update balance amount")
		return nil
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

		logger.Info().Msg("handled update currency amount")
		return nil
	}

	logger.Info().Msg("handled update balance")
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
	logger.Debug().Interface("opts", opts).Msg("got args")

	if opts.isBotCommand {
		err := h.messageService.SendMessage(&SendMessageOptions{
			ChatID: opts.chatID,
			Text:   "Enter balance amount:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		logger.Info().Msg("handled command input")
		return nil
	}

	price, err := money.NewFromString(opts.amount)
	if err != nil {
		logger.Error().Err(err).Msg("convert option amount to money type")

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
	logger.Debug().Interface("opts.balance.Amount", opts.balance.Amount).Msg("calculated balance amount")

	err = h.balanceStore.Update(ctx, opts.balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: opts.chatID,
		Text:   "Balance values successfully updated!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled case update balance amount")
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
	logger.Debug().Interface("opts", opts).Msg("got args")

	if opts.isBotCommand {
		err := h.messageService.SendMessage(&SendMessageOptions{
			ChatID: opts.chatID,
			Text:   "Enter balance currency:",
		})
		if err != nil {
			logger.Error().Err(err).Msg("send message")
			return fmt.Errorf("send message: %w", err)
		}

		logger.Info().Msg("handled command input")
		return nil
	}

	opts.balance.Currency = opts.currency

	err := h.balanceStore.Update(ctx, opts.balance)
	if err != nil {
		logger.Error().Err(err).Msg("update balance in store")
		return fmt.Errorf("update balance in store: %w", err)
	}

	err = h.messageService.SendMessage(&SendMessageOptions{
		ChatID: opts.chatID,
		Text:   "Balance currency successfully updated!",
	})
	if err != nil {
		logger.Error().Err(err).Msg("send message")
		return fmt.Errorf("send message: %w", err)
	}

	logger.Info().Msg("handled case update balance currency")
	return nil
}

func (h handlerService) HandleEventGetBalance(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

	balanceInfo, err := h.balanceService.GetBalanceInfo(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance info")
		return fmt.Errorf("get balance info: %w", err)
	}

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

	logger.Info().Msg("handled event get balance")
	return nil
}

func (h handlerService) HandleEventOperationCreate(ctx context.Context, eventName event, msg botMessage) error {
	logger := h.logger

	isBotCommand := IsBotCommand(msg.Message.Text) || IsBotCommand(msg.CallbackQuery.Data)

	if isBotCommand && eventName == createOperationEvent {
		err := h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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

	user, err := h.userService.GetUserByUsername(ctx, msg.GetUsername())
	if err != nil {
		logger.Error().Err(err).Msg("get user by username")
		return fmt.Errorf("get user by username: %w", err)
	}

	balance, err := h.balanceStore.Get(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balance from store")
		return fmt.Errorf("get balance from store: %w", err)
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

				logger.Info().Msg("categories not found")
				return nil
			}

			logger.Error().Err(err).Msg("get list of categories from store")
			return fmt.Errorf("get list of categories from store: %w", err)
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

		logger.Info().Msg("handled second command input")
		return nil
	}

	if msg.Message.Text != "" {
		category, err := h.categoryStore.GetByTitle(ctx, msg.Message.Text)
		if err != nil {
			logger.Error().Err(err).Msg("get category store")
			return fmt.Errorf("get category store: %w", err)
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

		err = h.operationStore.Create(ctx, &models.Operation{
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
		err := h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

	balanceInfo, err := h.balanceService.GetBalanceInfo(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balcne info by user id")
		return fmt.Errorf("get balance info by user id: %w", err)
	}

	operations, err := h.operationStore.GetAll(ctx, balanceInfo.ID, GetAllOperationsFilter{})
	if err != nil {
		logger.Error().Err(err).Msg("get all operations from store")
		return fmt.Errorf("get all operations from store: %w", err)
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

		logger.Info().Msg("operations not found")
		return nil
	}

	if operations[len(operations)-1].Amount == "" {
		operations[len(operations)-1].Amount = msg.Message.Text
		err = h.operationService.CreateOperation(ctx, CreateOperationOptions{
			UserID:    user.ID,
			Operation: &operations[len(operations)-1],
		})
		if err != nil {
			err = h.messageService.SendMessage(&SendMessageOptions{
				ChatID: msg.Message.Chat.ID,
				Text:   "Can't create operation!",
			})
			if err != nil {
				logger.Error().Err(err).Msg("send message")
				return fmt.Errorf("send message: %w", err)
			}

			logger.Error().Err(err).Msg("create operation in store")
			return fmt.Errorf("create operation in store: %w", err)
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

	logger.Info().Msg("handled event update operation amount")
	return nil
}

func (h handlerService) HandleEventGetOperationsHistory(ctx context.Context, msg botMessage) error {
	logger := h.logger
	logger.Debug().Interface("msg", msg).Msg("got args")

	if IsBotCommand(msg.Message.Text) {
		err := h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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

	user, err := h.userService.GetUserByUsername(ctx, msg.Message.From.Username)
	if err != nil {
		logger.Error().Err(err).Msg("get user from store")
		return fmt.Errorf("get user from store: %w", err)
	}

	balanceInfo, err := h.balanceService.GetBalanceInfo(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("get balcne info by user id")
		return fmt.Errorf("get balance info by user id: %w", err)
	}

	creationPeriod := models.GetCreationPeriodFromText(msg.Message.Text)
	if creationPeriod == nil {
		logger.Error().Msgf("message text is not creation period, text: %s", msg.Message.Text)
		return fmt.Errorf("message text is not creation period")
	}

	operations, err := h.operationStore.GetAll(ctx, balanceInfo.ID, GetAllOperationsFilter{
		CreationPeriod: creationPeriod,
	})
	if err != nil {
		logger.Error().Err(err).Msg("get all operations from store")
		return fmt.Errorf("get all operations from store: %w", err)
	}
	if operations == nil {
		logger.Info().Msg("operations not found")

		err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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

	err = h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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

	err := h.keyboardService.CreateKeyboard(&CreateKeyboardOptions{
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

	err := h.messageService.SendMessage(&SendMessageOptions{
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

	err := h.messageService.SendMessage(&SendMessageOptions{
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
