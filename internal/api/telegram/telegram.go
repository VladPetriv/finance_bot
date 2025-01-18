package telegram

import (
	"fmt"

	"github.com/VladPetriv/finance_bot/internal/service"
	"github.com/fasthttp/router"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
)

type telegramMessenger struct {
	api         *telego.Bot
	updatesType string
	srvAddr     string
	webhookURL  string
}

// Options represents options that required for creating new instance of telegram API.
type Options struct {
	// Token represents telegram bot token.
	Token string
	// UpdatesType represents a way we'll receive updates from Telegram. (webhook | polling)
	UpdatesType string

	// ServerAddress represents an address on which we'll start a server. (Required for webhook updates type)
	ServerAddress string
	// WebhookURL represents an url to which telegram will send updates. (Required for webhook updates type)
	WebhookURL string
}

// New creates a new instance of telegram API.
func New(opts Options) (*telegramMessenger, error) {
	bot, err := telego.NewBot(opts.Token, telego.WithDefaultLogger(false, true))
	if err != nil {
		return nil, fmt.Errorf("init bot instance: %w", err)
	}

	if opts.UpdatesType == "webhook" {
		err := bot.SetWebhook(&telego.SetWebhookParams{
			URL: opts.WebhookURL + "/bot",
		})
		if err != nil {
			return nil, fmt.Errorf("set webhook url: %w", err)
		}
	}

	return &telegramMessenger{
		api:         bot,
		updatesType: opts.UpdatesType,
		srvAddr:     opts.ServerAddress,
		webhookURL:  opts.WebhookURL,
	}, nil
}

func (t *telegramMessenger) ReadUpdates(result chan service.Message, errors chan error) {
	var (
		updates <-chan telego.Update
		err     error
	)

	switch t.updatesType {
	case "webhook":
		updates, err = t.api.UpdatesViaWebhook("/bot",
			telego.WithWebhookServer(telego.FastHTTPWebhookServer{
				Logger: t.api.Logger(),
				Server: &fasthttp.Server{},
				Router: router.New(),
			}),
		)
		if err != nil {
			errors <- fmt.Errorf("register webhook telegram updates receiver: %w", err)

			return
		}

		go func() {
			err := t.api.StartWebhook(t.srvAddr)
			if err != nil {
				errors <- fmt.Errorf("start webhook: %w", err)
			}
		}()
	case "polling":
		updates, err = t.api.UpdatesViaLongPolling(nil)
		if err != nil {
			errors <- fmt.Errorf("register long polling telegram updates receiver: %w", err)

			return
		}

	default:
		errors <- fmt.Errorf("unknown updates type: %s", t.updatesType)

		return
	}

	for update := range updates {
		result <- &TelegramUpdate{update: update}
	}
}

type TelegramUpdate struct {
	update telego.Update
}

func (t *TelegramUpdate) GetChatID() int {
	messageChatID := t.update.Message.Chat.ID
	callbackChatID := t.update.CallbackQuery.Message.Chat.ID

	if messageChatID != 0 {
		return int(messageChatID)
	}

	return int(callbackChatID)
}

func (t *TelegramUpdate) GetText() string {
	messageText := t.update.Message.Text
	callbackText := t.update.CallbackQuery.Data

	if messageText != "" {
		return messageText
	}

	return callbackText
}

func (t *TelegramUpdate) GetSenderName() string {
	messageSenderName := t.update.Message.From.FirstName
	callbackSenderName := t.update.CallbackQuery.From.FirstName

	if messageSenderName != "" {
		return messageSenderName
	}

	return callbackSenderName
}

func (t *telegramMessenger) Close() error {
	return t.api.StopWebhook()
}

func (t *telegramMessenger) SendMessage(chatID int, text string) error {
	return t.send(&sendOptions{
		chatID:  int64(chatID),
		message: text,
	})
}

func (t *telegramMessenger) SendWithKeyboard(opts service.SendWithKeyboardOptions) error {
	return t.send(&sendOptions{
		chatID:         int64(opts.ChatID),
		message:        opts.Message,
		keyboard:       opts.Keyboard,
		inlineKeyboard: opts.InlineKeyboard,
	})
}

type sendOptions struct {
	chatID  int64
	message string

	keyboard       []service.KeyboardRow
	inlineKeyboard []service.InlineKeyboardRow
}

func (t *telegramMessenger) send(opts *sendOptions) error {
	message := telegoutil.Message(telegoutil.ID(opts.chatID), opts.message)

	if len(opts.keyboard) != 0 {
		message = message.WithReplyMarkup(t.createKeyboard(opts.keyboard))
	}

	if len(opts.inlineKeyboard) != 0 {
		message = message.WithReplyMarkup(t.createInlineKeyboard(opts.inlineKeyboard))
	}

	_, err := t.api.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}

func (t *telegramMessenger) createKeyboard(rows []service.KeyboardRow) *telego.ReplyKeyboardMarkup {
	var convertedRows [][]telego.KeyboardButton

	for _, r := range rows {
		var buttons []telego.KeyboardButton

		for _, b := range r.Buttons {
			buttons = append(buttons, telegoutil.KeyboardButton(b))
		}

		convertedRows = append(convertedRows, buttons)
	}

	keyboard := telegoutil.Keyboard(convertedRows...).WithResizeKeyboard()

	return keyboard
}

func (t *telegramMessenger) createInlineKeyboard(rows []service.InlineKeyboardRow) *telego.InlineKeyboardMarkup {
	var convertedRows [][]telego.InlineKeyboardButton

	for _, r := range rows {
		var buttons []telego.InlineKeyboardButton

		for _, b := range r.Buttons {
			inlineKeyboardButton := telegoutil.
				InlineKeyboardButton(b.Text).
				WithCallbackData(b.Text)

			if b.Data != "" {
				inlineKeyboardButton = inlineKeyboardButton.WithCallbackData(b.Data)
			}

			buttons = append(buttons, inlineKeyboardButton)
		}

		convertedRows = append(convertedRows, buttons)
	}

	keyboard := telegoutil.InlineKeyboard(convertedRows...)

	return keyboard
}
