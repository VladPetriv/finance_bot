package bot

import (
	"encoding/json"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

type bot struct {
	token string
}

var _ Bot = (*bot)(nil)

type botAPI struct {
	api *telego.Bot
}

var _ API = (*botAPI)(nil)

// NewTelegramBot creates a new instance of telegram bot.
func NewTelegramBot(token string) *bot {
	return &bot{
		token: token,
	}
}

func (b bot) NewAPI() (API, error) {
	tgBot, err := telego.NewBot(b.token, telego.WithDefaultDebugLogger())
	if err != nil {
		return nil, err
	}

	return &botAPI{
		api: tgBot,
	}, nil
}

func (b botAPI) ReadUpdates(result chan []byte, errors chan error) {
	updates, err := b.api.UpdatesViaLongPolling(nil)
	if err != nil {
		errors <- err
	}

	defer b.api.StopLongPolling()

	for u := range updates {
		updatedData, err := json.Marshal(u)
		if err != nil {
			errors <- err

			continue
		}

		result <- updatedData
	}
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
