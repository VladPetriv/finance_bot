package service

import (
	"context"

	"github.com/VladPetriv/finance_bot/pkg/errs"
	"github.com/VladPetriv/finance_bot/pkg/money"
)

// APIs represents structure containing all the APIs that the application uses.
type APIs struct {
	Messenger         Messenger
	Prompter          Prompter
	CurrencyExchanger CurrencyExchanger
}

// Messenger handles messaging operations between the application and messaging platform.
type Messenger interface {
	// ReadUpdates retrieves new incoming updates/messages from the messaging platform.
	ReadUpdates(result chan Message, errors chan error)
	// SendMessage sends a text message to the specified chat.
	SendMessage(chatID int, text string) error
	// SendWithKeyboard sends a message with an attached keyboard (inline or reply).
	SendWithKeyboard(opts SendWithKeyboardOptions) error

	// Close closes the underlying connection to the messaging platform.
	Close() error
}

// SendWithKeyboardOptions represents options for sending a message with a keyboard.
type SendWithKeyboardOptions struct {
	ChatID                  int
	Message                 string
	FormatMessageInMarkDown bool
	Keyboard                []KeyboardRow
	InlineKeyboard          []InlineKeyboardRow
}

// KeyboardRow represents keyboard row with buttons.
type KeyboardRow struct {
	Buttons []string
}

// InlineKeyboardRow represents inline keyboard row with buttons.
type InlineKeyboardRow struct {
	Buttons []InlineKeyboardButton
}

// InlineKeyboardButton represents an inline keyboard button with text and data.
type InlineKeyboardButton struct {
	Text string
	Data string
}

// Message represents a message that was received from the messaging platform.
type Message interface {
	// GetChatID returns the ID of the chat the message was sent to.
	GetChatID() int
	// GetMessageID returns the ID of the message.
	GetMessageID() int
	// GetInlineMessageID returns the ID of the inline message.
	GetInlineMessageID() string
	// GetText returns the text content of the message.
	GetText() string
	// GetSenderName returns the name of the user who sent the message.
	GetSenderName() string
}

// CurrencyExchanger handles currency exchange rates and supported currencies listing
type CurrencyExchanger interface {
	// FetchCurrencies returns a list of available currencies.
	FetchCurrencies() ([]Currency, error)
	// GetExchangeRate returns the exchange rate for the specified currency.
	GetExchangeRate(baseCurrency, targetCurrency string) (*money.Money, error)
}

// Currency represents a structure that contains currency name, code and symbol.
type Currency struct {
	Name   string
	Code   string
	Symbol string
}

// ErrCurrencyExchangeRateNotFound happens when the CurrencyExchanger cannot find the exchange rate for the specified currency.
var ErrCurrencyExchangeRateNotFound = errs.New("currency exchange rate not found")

// Prompter executes prompts and returns generated responses
type Prompter interface {
	// Execute processes the prompt and returns the response
	Execute(ctx context.Context, prompt string) (string, error)
}
