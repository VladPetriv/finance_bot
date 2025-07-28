package telegram

import (
	"github.com/mymmrac/telego"
)

// Update represents the update received from the Telegram.
type Update struct {
	update telego.Update
}

// GetChatID returns the ID of the chat the message was sent to.
func (t *Update) GetChatID() int {
	var chatID int

	if t.update.Message != nil {
		chatID = int(t.update.Message.Chat.ID)
	}
	if t.update.CallbackQuery != nil {
		chatID = int(t.update.CallbackQuery.Message.Chat.ID)
	}

	return chatID
}

// GetMessageID returns the ID of the message.
func (t *Update) GetMessageID() int {
	var messageID int

	if t.update.Message != nil {
		messageID = t.update.Message.MessageID
	}
	if t.update.CallbackQuery != nil {
		messageID = t.update.CallbackQuery.Message.MessageID
	}

	return messageID
}

// GetInlineMessageID returns the inline message ID of the callback query.
func (t *Update) GetInlineMessageID() string {
	if t.update.CallbackQuery != nil {
		return t.update.CallbackQuery.InlineMessageID
	}

	return ""
}

// GetText returns the text content of the message or callback data.
func (t *Update) GetText() string {
	var text string

	if t.update.Message != nil {
		text = t.update.Message.Text
	}
	if t.update.CallbackQuery != nil {
		text = t.update.CallbackQuery.Data
	}

	return text
}

// GetSenderName returns the name of the user who sent the message.
func (t *Update) GetSenderName() string {
	var senderName string

	if t.update.Message != nil {
		senderName = t.update.Message.From.Username
	}
	if t.update.CallbackQuery != nil {
		senderName = t.update.CallbackQuery.From.Username
	}

	return senderName
}
