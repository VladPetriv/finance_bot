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
	updatedMessage := opts.UpdatedMessage
	if opts.FormatMessageInMarkDown {
		// Unescape markdown symbols.
		message = strings.ReplaceAll(message, "(", `\(`)
		message = strings.ReplaceAll(message, ")", `\)`)
		message = strings.ReplaceAll(message, "!", `\!`)
		message = strings.ReplaceAll(message, "-", `\-`)
		message = strings.ReplaceAll(message, "+", `\+`)
		message = strings.ReplaceAll(message, ".", `\.`)

		updatedMessage = strings.ReplaceAll(updatedMessage, "(", `\(`)
		updatedMessage = strings.ReplaceAll(updatedMessage, ")", `\)`)
		updatedMessage = strings.ReplaceAll(updatedMessage, "!", `\!`)
		updatedMessage = strings.ReplaceAll(updatedMessage, "-", `\-`)
		updatedMessage = strings.ReplaceAll(updatedMessage, "+", `\+`)
		updatedMessage = strings.ReplaceAll(updatedMessage, ".", `\.`)
	}

	return t.send(&sendOptions{
		chatID:                int64(opts.ChatID),
		messageID:             opts.MessageID,
		inlineMessageID:       opts.InlineMessageID,
		message:               message,
		formatInMarkdown:      opts.FormatMessageInMarkDown,
		keyboard:              opts.Keyboard,
		inlineKeyboard:        opts.InlineKeyboard,
		updatedInlineKeyboard: opts.UpdatedInlineKeyboard,
		updatedMessage:        updatedMessage,
	})
}

type sendOptions struct {
	chatID          int64
	messageID       int
	inlineMessageID string

	message          string
	formatInMarkdown bool

	keyboard              []service.KeyboardRow
	inlineKeyboard        []service.InlineKeyboardRow
	updatedInlineKeyboard []service.InlineKeyboardRow
	updatedMessage        string
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

	if len(opts.updatedMessage) > 0 {
		var inlineKeyboard *telego.InlineKeyboardMarkup

		if len(opts.updatedInlineKeyboard) > 0 {
			inlineKeyboards := t.createInlineKeyboard(opts.updatedInlineKeyboard)
			inlineKeyboard = inlineKeyboards[0]
		}

		editMessageParams := &telego.EditMessageTextParams{
			ChatID: telego.ChatID{
				ID: opts.chatID,
			},
			MessageID:       opts.messageID,
			InlineMessageID: opts.inlineMessageID,
			Text:            opts.updatedMessage,
			ReplyMarkup:     inlineKeyboard,
		}
		if opts.formatInMarkdown {
			editMessageParams.ParseMode = markdownFormat
		}

		_, err := t.api.EditMessageText(editMessageParams)
		if err != nil {
			return fmt.Errorf("edit message text: %w", err)
		}

		return nil
	}

	if len(opts.updatedInlineKeyboard) > 0 {
		inlineKeyboard := t.createInlineKeyboard(opts.updatedInlineKeyboard)

		_, err := t.api.EditMessageReplyMarkup(&telego.EditMessageReplyMarkupParams{
			ChatID: telego.ChatID{
				ID: opts.chatID,
			},
			MessageID:       opts.messageID,
			InlineMessageID: opts.inlineMessageID,
			ReplyMarkup:     inlineKeyboard[0],
		})
		if err != nil {
			return fmt.Errorf("edit message reply markup: %w", err)
		}

		return nil
	}

	_, err := t.api.SendMessage(message)
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
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

const maxButtonsPerMessage = 100

func (t *telegramMessenger) createInlineKeyboard(rows []service.InlineKeyboardRow) []*telego.InlineKeyboardMarkup {
	convertedRows := make([][]telego.InlineKeyboardButton, 0)

	var totalButtonsCount int

	for _, r := range rows {
		var buttons []telego.InlineKeyboardButton

		for _, b := range r.Buttons {
			totalButtonsCount++

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

	if totalButtonsCount <= maxButtonsPerMessage {
		return []*telego.InlineKeyboardMarkup{telegoutil.InlineKeyboard(convertedRows...)}
	}

	return splitInlineKeyboardRows(convertedRows, maxButtonsPerMessage)
}

func splitInlineKeyboardRows(convertedRows [][]telego.InlineKeyboardButton, maxButtonsPerMessage int) []*telego.InlineKeyboardMarkup {
	var (
		buttonsCount        int
		lastProcessedRowIdx int
	)

	result := make([]*telego.InlineKeyboardMarkup, 0, 2)

	for rowIdx, row := range convertedRows {
		for btnIdx := range row {
			if buttonsCount == maxButtonsPerMessage {
				splitIndex := rowIdx

				// Ensure the split doesn't occur in the middle of a row
				if btnIdx > 0 {
					splitIndex--
				}

				result = append(result, telegoutil.InlineKeyboard(convertedRows[lastProcessedRowIdx:splitIndex]...))

				// Reset counters
				buttonsCount = 0
				lastProcessedRowIdx = splitIndex
			}

			buttonsCount++
		}
	}

	// Append the remaining buttons
	if lastProcessedRowIdx < len(convertedRows) {
		result = append(result, telegoutil.InlineKeyboard(convertedRows[lastProcessedRowIdx:]...))
	}

	return result
}
