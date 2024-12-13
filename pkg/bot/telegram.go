package bot

import (
	"encoding/json"
	"fmt"

	"github.com/fasthttp/router"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
)

type bot struct {
	token      string
	webhookURL string
}

var _ Bot = (*bot)(nil)

type botAPI struct {
	api        *telego.Bot
	webhookURL string
}

var _ API = (*botAPI)(nil)

// NewTelegramBot creates a new instance of telegram bot.
func NewTelegramBot(token, webhookURL string) Bot {
	return &bot{
		token:      token,
		webhookURL: webhookURL,
	}
}

func (b bot) NewAPI() (API, error) {
	tgBot, err := telego.NewBot(b.token, telego.WithDefaultLogger(false, true))
	if err != nil {
		return nil, err
	}

	err = tgBot.SetWebhook(&telego.SetWebhookParams{
		URL: b.webhookURL + "/bot",
	})
	if err != nil {
		return nil, err
	}

	return &botAPI{
		api:        tgBot,
		webhookURL: b.webhookURL,
	}, nil
}

func (b botAPI) ReadUpdates(result chan []byte, errors chan error) {
	updates, err := b.api.UpdatesViaWebhook("/bot",
		telego.WithWebhookServer(telego.FastHTTPWebhookServer{
			Logger: b.api.Logger(),
			Server: &fasthttp.Server{},
			Router: router.New(),
		}),
	)
	if err != nil {
		errors <- fmt.Errorf("register teelgram updates receiver: %w", err)

		return
	}

	go func() {
		err := b.api.StartWebhook(b.webhookURL)
		fmt.Printf("err: %v\n", err)
	}()

	for update := range updates {
		rawUpdateData, err := json.Marshal(update)
		if err != nil {
			errors <- fmt.Errorf("marshal telegram update: %w", err)

			continue
		}

		result <- rawUpdateData
	}
}

func (b botAPI) Close() error {
	return b.api.StopWebhook()
}

func (b botAPI) Send(opts *SendOptions) error {
	message := telegoutil.Message(telegoutil.ID(opts.ChatID), opts.Message)

	if opts.Keyboard != nil {
		message = message.WithReplyMarkup(b.createKeyboard(opts.Keyboard))
	}

	if opts.InlineKeyboard != nil {
		message = message.WithReplyMarkup(b.createInlineKeyboard(opts.InlineKeyboard))
	}

	_, err := b.api.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}

func (b botAPI) createKeyboard(rows []KeyboardRow) *telego.ReplyKeyboardMarkup {
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

func (b botAPI) createInlineKeyboard(rows []KeyboardRow) *telego.InlineKeyboardMarkup {
	var convertedRows [][]telego.InlineKeyboardButton

	for _, r := range rows {
		var buttons []telego.InlineKeyboardButton

		for _, b := range r.Buttons {
			buttons = append(buttons, telegoutil.InlineKeyboardButton(b).WithCallbackData(b))
		}

		convertedRows = append(convertedRows, buttons)
	}

	keyboard := telegoutil.InlineKeyboard(convertedRows...)

	return keyboard
}
