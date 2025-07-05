package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) showMainMenu(chatID int64) {
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

func (b *Bot) showTrendPlaceholder(chatID int64) {
	err := b.Send(tgbotapi.NewMessage(chatID, "Trends feature is not available yet."))
	if err != nil {
		b.logger.Error(err)
	}
}

func (b *Bot) showDebtMenu(chatID int64) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Add Debt"),
			tgbotapi.NewKeyboardButton("List Debts"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Pay Debt"),
			tgbotapi.NewKeyboardButton("Delete Debt"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Update Debt"),
			tgbotapi.NewKeyboardButton("Back"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Debt Menu:")
	msg.ReplyMarkup = keyboard
	err := b.Send(msg)
	if err != nil {
		b.logger.Error(err)
	}
}
