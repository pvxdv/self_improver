package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
	"time"
)

func (b *Bot) handleAuth(chatID int64, input string, state *UserState) {
	b.logger.Infof("Auth attempt. ChatID: %d, Input: %s", chatID, input)

	if strings.TrimSpace(input) == "/start" {
		state.AuthPassed = false
		state.ExpectedAction = ""
		state.TempData = make(map[string]interface{})
		b.showAuthMenu(chatID)
		return
	}

	if state.AuthPassed {
		b.showMainMenu(chatID)
		return
	}

	switch state.ExpectedAction {
	case "await_password":
		if !b.IsPassValid(input) {
			b.logger.Warnf("Invalid password attempt from chat %d", chatID)
			b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" invalid password. Please try again:"))
			return
		}

		state.AuthPassed = true
		state.ExpectedAction = ""
		b.logger.Infof("User %d successfully authorized", chatID)
		b.Send(tgbotapi.NewMessage(chatID, emojiSuccess+" authorization successful!"))
		time.Sleep(1 * time.Second)
		b.showMainMenu(chatID)

	default:
		b.logger.Warnf("Unauthorized access attempt from chat %d", chatID)
		msg := tgbotapi.NewMessage(chatID, emojiAuth+" please authorize first")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(emojiAuth+" authorize", CallbackAuthorize),
			),
		)
		b.Send(msg)
	}
}
