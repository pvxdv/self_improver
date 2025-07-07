package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
	"time"
)

func (b *Bot) formatAmount(amount int64) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	s := fmt.Sprintf("%d", amount)
	n := len(s)
	var result strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteByte(s[i])
	}
	if neg {
		return "- " + result.String() + "₽"
	}
	return result.String() + "₽"
}

func (b *Bot) formatDate(date *time.Time) string {
	if date == nil {
		return "Not set"
	}
	return date.Format("2006-01-02")
}

func (b *Bot) showYearPicker(chatID int64, debtID int64) {
	var rows [][]tgbotapi.InlineKeyboardButton
	for y := time.Now().Year(); y <= time.Now().Year()+5; y++ {
		year := fmt.Sprintf("%d", y)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(year, CallbackWithData(CallbackSetYear, debtID, int64(y))),
		))
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiCalendar+" Select year:"))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = &keyboard
	b.Send(msg)

	state := b.getUserState(chatID)
	state.TempData["debt_id"] = debtID
	state.ExpectedAction = "select_year"
}

func (b *Bot) showMonthPicker(chatID int64, debtID int64, year int64) {
	months := []struct {
		Name string
		Num  int64
	}{
		{"Jan", 1}, {"Feb", 2}, {"Mar", 3}, {"Apr", 4},
		{"May", 5}, {"Jun", 6}, {"Jul", 7}, {"Aug", 8},
		{"Sep", 9}, {"Oct", 10}, {"Nov", 11}, {"Dec", 12},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(months); i += 3 {
		end := i + 3
		if end > len(months) {
			end = len(months)
		}

		var row []tgbotapi.InlineKeyboardButton
		for _, m := range months[i:end] {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				m.Name,
				CallbackWithData(CallbackSetMonth, debtID, year, m.Num),
			))
		}
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiCalendar+" Select month for %d:", year))
	msg.ReplyMarkup = &keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Warnw("Failed to send month picker", "error", err, "chat_id", chatID, "debt_id", debtID)
		return
	}

	b.logger.Debug("Showing month picker",
		"chat_id", chatID,
		"debt_id", debtID,
		"year", year,
		"button_count", len(months),
	)

	state := b.getUserState(chatID)
	state.TempData["year"] = year
	state.ExpectedAction = "select_month"
}

func (b *Bot) showDayPicker(chatID int64, debtID int64, year int64, month int64) {
	monthTime := time.Date(int(year), time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := monthTime.AddDate(0, 1, -1).Day()

	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for d := 1; d <= daysInMonth; d++ {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d", d),
			CallbackWithData(CallbackSetDay, debtID, year, month, int64(d)),
		))

		if len(row) == 7 || d == daysInMonth {
			rows = append(rows, row)
			row = nil
		}
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiCalendar+" Select day for %d-%d:", year, month))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = &keyboard
	b.Send(msg)

	b.logger.Debug("Showing days picker",
		"chat_id", chatID,
		"debt_id", debtID,
		"year", year,
		"button_count", daysInMonth,
	)

	state := b.getUserState(chatID)
	state.TempData["year"] = year
	state.TempData["month"] = month
	state.ExpectedAction = "select_day"
}
