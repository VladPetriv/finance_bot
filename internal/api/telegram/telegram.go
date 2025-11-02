package telegram

import (
	"fmt"
	"strings"

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
		// Answer on callback query if it exists to avoid spam from telegram if user closes the application.
		if update.CallbackQuery != nil {
			err := t.api.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				ShowAlert:       false,
			})
			if err != nil {
				errors <- fmt.Errorf("answer callback query: %w", err)
			}
		}

		result <- &Update{update: update}
	}
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
	message := opts.Message
	if opts.FormatMessageInMarkDown {
		message = unescapeMarkdownSymbols(message)
	}

	return t.send(&sendOptions{
		chatID:           int64(opts.ChatID),
		message:          message,
		formatInMarkdown: opts.FormatMessageInMarkDown,
		keyboard:         opts.Keyboard,
		inlineKeyboard:   opts.InlineKeyboard,
	})
}

func (t *telegramMessenger) UpdateMessage(opts service.UpdateMessageOptions) error {
	// NOTE: Telegram does not support direct editing of messages with row keyboards.
	// Instead delete the entire message and send a new one with the updated message and keyboard
	if len(opts.UpdatedKeyboard) != 0 {
		err := t.api.DeleteMessage(&telego.DeleteMessageParams{
			ChatID: telego.ChatID{
				ID: int64(opts.ChatID),
			},
			MessageID: opts.MessageID,
		})
		if err != nil {
			return fmt.Errorf("delete message: %w", err)
		}

		if opts.FormatMessageInMarkDown {
			opts.UpdatedMessage = unescapeMarkdownSymbols(opts.UpdatedMessage)
		}

		return t.send(&sendOptions{
			chatID:           int64(opts.ChatID),
			message:          opts.UpdatedMessage,
			formatInMarkdown: opts.FormatMessageInMarkDown,
			keyboard:         opts.UpdatedKeyboard,
		})
	}

	editMessageParams := &telego.EditMessageTextParams{
		ChatID: telego.ChatID{
			ID: int64(opts.ChatID),
		},
		MessageID:       opts.MessageID,
		InlineMessageID: opts.InlineMessageID,
		Text:            opts.UpdatedMessage,
	}

	if len(opts.UpdatedInlineKeyboard) > 0 {
		inlineKeyboards := t.createInlineKeyboard(opts.UpdatedInlineKeyboard)
		editMessageParams.ReplyMarkup = inlineKeyboards[0]
	}

	if opts.FormatMessageInMarkDown {
		editMessageParams.ParseMode = markdownFormat
		editMessageParams.Text = unescapeMarkdownSymbols(editMessageParams.Text)
		editMessageParams = editMessageParams.WithParseMode(markdownFormat)
	}

	_, err := t.api.EditMessageText(editMessageParams)
	if err != nil {
		return fmt.Errorf("edit message text: %w", err)
	}

	return nil
}

func unescapeMarkdownSymbols(message string) string {
	message = strings.ReplaceAll(message, "(", `\(`)
	message = strings.ReplaceAll(message, ")", `\)`)
	message = strings.ReplaceAll(message, "!", `\!`)
	message = strings.ReplaceAll(message, "-", `\-`)
	message = strings.ReplaceAll(message, "+", `\+`)
	message = strings.ReplaceAll(message, ".", `\.`)
	return message
}

type sendOptions struct {
	chatID int64

	message          string
	formatInMarkdown bool

	keyboard       []service.KeyboardRow
	inlineKeyboard []service.InlineKeyboardRow
}

const (
	markdownFormat = "MarkdownV2"
	emptyMessage   = "ㅤ"
)

func (t *telegramMessenger) send(opts *sendOptions) error {
	message := telegoutil.
		Message(telegoutil.ID(opts.chatID), opts.message)

	if opts.formatInMarkdown {
		message = message.WithParseMode(markdownFormat)
	}

	if len(opts.keyboard) != 0 {
		message = message.WithReplyMarkup(t.createKeyboard(opts.keyboard))
	}

	if len(opts.inlineKeyboard) != 0 {
		inlineKeyboards := t.createInlineKeyboard(opts.inlineKeyboard)

		if len(inlineKeyboards) > 1 {
			message := message.WithReplyMarkup(inlineKeyboards[0])
			_, err := t.api.SendMessage(message)
			if err != nil {
				return fmt.Errorf("send telegram message: %w", err)
			}

			for _, inlineKeyboard := range inlineKeyboards[1:] {
				message := telegoutil.
					Message(telegoutil.ID(opts.chatID), emptyMessage).
					WithReplyMarkup(inlineKeyboard)

				_, err := t.api.SendMessage(message)
				if err != nil {
					return fmt.Errorf("send telegram message: %w", err)
				}
			}

			return nil
		}

		message = message.WithReplyMarkup(inlineKeyboards[0])
	}

	_, err := t.api.SendMessage(message)
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}

	return nil
}
