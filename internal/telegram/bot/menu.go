package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sort"
	"strings"
	"time"
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
	err := b.Send(msg)
	b.logger.Warnf("Failed to send message: %+v", err)
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
	err := b.Send(msg)
	b.logger.Warnf("Failed to send message: %+v", err)
}

func (b *Bot) showDebtMenu(chatID int64) {
	b.logger.Debugf("Showing debt menu for chat %d", chatID)

	debts, err := b.serviceDebt.GetAmountDebt(b.ctx)
	if err != nil {
		b.logger.Errorf("Error fetching debts: %v", err)
		err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error loading debts. Please try later."))
		b.logger.Warnf("Failed to send message: %+v", err)
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
		err = b.Send(msg)
		b.logger.Warnf("Failed to send message: %+v", err)
		return
	}

	sort.Slice(debts.Details, func(i, j int) bool {
		a, b := debts.Details[i], debts.Details[j]

		if a.ReturnDate == nil && b.ReturnDate == nil {
			return a.ID > b.ID
		}

		if a.ReturnDate == nil {
			return false
		}
		if b.ReturnDate == nil {
			return true
		}

		return a.ReturnDate.After(*b.ReturnDate)
	})

	for _, debt := range debts.Details {
		var (
			messageText  strings.Builder
			keyboardRows [][]tgbotapi.InlineKeyboardButton
		)

		daysRemaining := 0
		if debt.ReturnDate != nil {
			d1 := debt.ReturnDate.UTC()
			d2 := time.Now().UTC()

			dif := d1.Sub(d2)

			days := int(dif.Hours() / 24)
			if days > 0 {
				daysRemaining = days
			}
		}

		messageText.WriteString(fmt.Sprintf(
			"%s: %s\n%s: %s %s: %s %s: %dd\n\n ",
			emojiInfo,
			debt.Description,
			emojiMoney,
			b.formatAmount(debt.Amount),
			emojiCalendar,
			b.formatDate(debt.ReturnDate),
			b.getDeadlineIcon(daysRemaining),
			daysRemaining,
		))

		keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("pay", CallbackWithData(CallbackPayDebt, debt.ID)),
			tgbotapi.NewInlineKeyboardButtonData("edit", CallbackWithData(CallbackEditDebt, debt.ID)),
			tgbotapi.NewInlineKeyboardButtonData("delete", CallbackWithData(CallbackDeleteDebt, debt.ID)),
		))

		msg := tgbotapi.NewMessage(chatID, messageText.String())
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

		err = b.Send(msg)
		b.logger.Warnf("Failed to send message: %+v", err)
	}

	var (
		messageText  strings.Builder
		keyboardRows [][]tgbotapi.InlineKeyboardButton
	)

	messageText.WriteString(fmt.Sprintf(
		"Total debt:%s\n\n",
		b.formatAmount(debts.Amount)))

	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("new", CallbackAddDebt),
		tgbotapi.NewInlineKeyboardButtonData("home", CallbackMainMenu),
	))

	msg := tgbotapi.NewMessage(chatID, messageText.String())
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	err = b.Send(msg)
	b.logger.Warnf("Failed to send message: %+v", err)
}
