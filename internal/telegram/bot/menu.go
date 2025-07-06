package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sort"
	"strings"
)

func (b *Bot) showMainMenu(chatID int64) {
	b.logger.Debugf("Showing main menu for chat %d", chatID)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(emojiTrends+" trends", CallbackTrends),
			tgbotapi.NewInlineKeyboardButtonData(emojiDebts+" debts", CallbackDebtMenu),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Choose module:")
	msg.ReplyMarkup = keyboard
	b.Send(msg)
}

func (b *Bot) showAuthMenu(chatID int64) {
	b.logger.Debugf("Showing auth menu for chat %d", chatID)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(emojiAuth+" authorize", CallbackAuthorize),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Welcome to the bot!\nPlease authorize to continue.")
	msg.ReplyMarkup = keyboard
	b.Send(msg)
}

func (b *Bot) showDebtMenu(chatID int64) {
	b.logger.Debugf("Showing debt menu for chat %d", chatID)

	debts, err := b.serviceDebt.GetAmountDebt(b.ctx)
	if err != nil {
		b.logger.Errorf("Error fetching debts: %v", err)
		b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error loading debts. Please try later."))
		return
	}

	b.logger.Debugf("Fetched %d debts", len(debts.Details))

	if len(debts.Details) == 0 {
		msg := tgbotapi.NewMessage(chatID, emojiSuccess+"no debts found")
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(emojiActionAdd+" add Debt", CallbackAddDebt),
				tgbotapi.NewInlineKeyboardButtonData(emojiActionBack+" home", CallbackMainMenu),
			),
		)
		msg.ReplyMarkup = keyboard
		b.Send(msg)
		return
	}

	sort.Slice(debts.Details, func(i, j int) bool {
		a, b := debts.Details[i], debts.Details[j]

		if a.ReturnDate == nil && b.ReturnDate == nil {
			return a.ID < b.ID
		}

		if a.ReturnDate == nil {
			return false
		}
		if b.ReturnDate == nil {
			return true
		}

		return a.ReturnDate.Before(*b.ReturnDate)
	})

	var (
		messageText  strings.Builder
		keyboardRows [][]tgbotapi.InlineKeyboardButton
	)

	messageText.WriteString(fmt.Sprintf(
		"Total debt: %s:%s\n\n",
		emojiMoney,
		formatAmount(debts.Amount)))

	for _, debt := range debts.Details {
		messageText.WriteString(fmt.Sprintf(
			"%s:%d\n%s:%s\n%s:%s\n%s:%s\n\n",
			emojiId,
			debt.ID,
			emojiInfo,
			debt.Description,
			emojiMoney,
			formatAmount(debt.Amount),
			emojiCalendar,
			formatDate(debt.ReturnDate),
		))

		keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("(%s:%d) PAY", emojiId, debt.ID), CallbackWithData(CallbackPayDebt, debt.ID)),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("(%s:%d) EDIT", emojiId, debt.ID), CallbackWithData(CallbackEditDebt, debt.ID)),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("(%s:%d) DELETE", emojiId, debt.ID), CallbackWithData(CallbackDeleteDebt, debt.ID)),
		))
	}

	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(emojiActionAdd+" add Debt", CallbackAddDebt),
		tgbotapi.NewInlineKeyboardButtonData(emojiActionBack+" home", CallbackMainMenu),
	))

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	b.Send(msg)
}
