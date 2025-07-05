package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleAuth(b *Bot, chatID int64, input string, state *UserState) {
	if b.IsPassValid(input) {
		state.AuthPassed = true
		SendMainMenu(b, chatID)
	} else {
		msg := tgbotapi.NewMessage(chatID, "Неверный пароль. Попробуйте снова.")
		err := b.Send(msg)
		if err != nil {
			b.logger.Error(err)
		}
	}
}

func SendMainMenu(b *Bot, chatID int64) {
	msgText := "Choose an action:"
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Trends"),
			tgbotapi.NewKeyboardButton("Debts"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ReplyMarkup = keyboard
	err := b.Send(msg)
	if err != nil {
		b.logger.Error(err)
	}
}
