package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackMainMenu          = "main_menu"
	CallbackDebtMenu          = "debt_menu"
	CallbackAddDebt           = "add_debt"
	CallbackPayDebt           = "pay_debt"
	CallbackDeleteDebt        = "delete_debt"
	CallbackAuthorize         = "authorize"
	CallbackTrends            = "trends"
	CallbackEditDebt          = "edit_debt"
	CallbackUpdateAmount      = "update_amount"
	CallbackUpdateDescription = "update_description"
	CallbackConfirmDelete     = "confirm_delete"
	CallbackCancelDelete      = "cancel_delete"
	CallbackUpdateReturnDate  = "update_return_date"

	CalendarPrefix   = "calendar_"
	CallbackSetYear  = "calendar_set_year"
	CallbackSetMonth = "calendar_set_month"
	CallbackSetDay   = "calendar_set_day"
)

//func CallbackWithData(prefix string, id int64) string {
//	return fmt.Sprintf("%s:%d", prefix, id)
//}

func CallbackWithData(prefix string, ids ...int64) string {
	data := prefix
	for _, id := range ids {
		data += fmt.Sprintf(":%d", id)
	}
	return data
}

func ParseCallbackData(data string) (string, int64, error) {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid callback data: %s", data)
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid callback data: %s", data)
	}

	return parts[0], id, err
}

func ParseCalendarCallback(data string) (string, []int64, error) {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid calendar callback")
	}

	var nums []int64
	for _, part := range parts[1:] {
		num, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return "", nil, err
		}
		nums = append(nums, num)
	}

	return parts[0], nums, nil
}

func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	callbackData := callback.Data

	if _, err := b.api.Request(tgbotapi.NewCallback(callback.ID, "")); err != nil {
		b.logger.Warnw("Failed to answer callback query", "error", err)
	}

	state := b.getUserState(chatID)
	b.logger.Debug("Processing callback:", callbackData)

	if callback.Data != CallbackAuthorize && !state.AuthPassed {
		b.handleAuth(chatID, "", state)
		return
	}

	switch callbackData {
	case CallbackMainMenu:
		b.showMainMenu(chatID)
		return
	case CallbackDebtMenu:
		b.showDebtMenu(chatID)
		return
	case CallbackAddDebt:
		state.ExpectedAction = "add_debt_amount"
		state.TempData = make(map[string]interface{})

		err := b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" enter debt amount:"))
		b.logger.Warnf("Failed to send message: %+v", err)
		return
	case CallbackAuthorize:
		state.ExpectedAction = "await_password"

		err := b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" enter your password:"))
		b.logger.Warnf("Failed to send message: %+v", err)
		return
	case CallbackTrends:
		err := b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" feature is not implemented yet"))
		b.logger.Warnf("Failed to send message: %+v", err)
		return
	case CallbackCancelDelete:
		err := b.Send(tgbotapi.NewMessage(chatID, emojiActionDelete+" deletion cancelled"))
		b.logger.Warnf("Failed to send message: %+v", err)
		b.showDebtMenu(chatID)
		return
	default:

	}

	if strings.HasPrefix(callbackData, CalendarPrefix) {
		b.handleCalendarCallback(callbackData, chatID, state)
		return
	}

	prefix, id, err := ParseCallbackData(callbackData)
	if err != nil {
		err = b.Send(tgbotapi.NewMessage(chatID, emojiWarning+" invalid command format. Please try again."))
		b.logger.Warnf("Failed to send message: %+v", err)
		return
	}

	switch prefix {
	case CallbackEditDebt:
		state.TempData["debt_id"] = id
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(emojiMoney+" amount", fmt.Sprintf("%s:%d", CallbackUpdateAmount, id)),
				tgbotapi.NewInlineKeyboardButtonData(emojiInfo+" description", fmt.Sprintf("%s:%d", CallbackUpdateDescription, id)),
				tgbotapi.NewInlineKeyboardButtonData(emojiCalendar+" return Date", fmt.Sprintf("%s:%d", CallbackUpdateReturnDate, id)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(emojiActionBack+" cancel", CallbackDebtMenu),
			),
		)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiInfo+" what do you want to update for Debt #%d?", id))
		msg.ReplyMarkup = keyboard
		err = b.Send(msg)
		b.logger.Warnf("Failed to send message: %+v", err)

	case CallbackUpdateAmount:
		state.ExpectedAction = "update_debt_amount"
		state.TempData["debt_id"] = id

		err = b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiInfo+" enter new amount for Debt #%d:", id)))
		b.logger.Warnf("Failed to send message: %+v", err)

	case CallbackUpdateDescription:
		state.ExpectedAction = "update_debt_description"
		state.TempData["debt_id"] = id

		err = b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiInfo+" enter new description for Debt #%d:", id)))
		b.logger.Warnf("Failed to send message: %+v", err)

	case CallbackPayDebt:
		state.ExpectedAction = "pay_debt_amount"
		state.TempData["debt_id"] = id

		err = b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" enter payment amount:"))
		b.logger.Warnf("Failed to send message: %+v", err)

	case CallbackDeleteDebt:
		debt, err := b.serviceDebt.GetDebt(b.ctx, id)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+"️ error: Debt not found"))
			b.logger.Warnf("Failed to send message: %+v", err)
			return
		}

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			emojiWarning+"️ Confirm Deletion, this cannot be undone!\n\n"+
				"To delete:\n\n"+
				emojiPoint+" "+emojiId+": %d\n"+
				emojiPoint+" "+emojiInfo+": %s\n"+
				emojiPoint+" "+emojiMoney+": %s\n\n",
			id, debt.Description, b.formatAmount(debt.Amount)))

		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(emojiActionDelete, fmt.Sprintf("%s:%d", CallbackConfirmDelete, id)),
				tgbotapi.NewInlineKeyboardButtonData(emojiActionBack, CallbackCancelDelete),
			),
		)

		err = b.Send(msg)
		b.logger.Warnf("Failed to send message: %+v", err)

	case CallbackConfirmDelete:
		if err := b.serviceDebt.DeleteDebt(b.ctx, id); err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" failed to delete debt: "+err.Error()))
			b.logger.Warnf("Failed to send message: %+v", err)
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiSuccess+" debt deleted successfully"))
			b.logger.Warnf("Failed to send message: %+v", err)
		}

		b.showDebtMenu(chatID)

	case CallbackUpdateReturnDate:
		b.showYearPicker(chatID, id)

	default:
		b.logger.Warnf("Unknown callback prefix: %s", prefix)

		err = b.Send(tgbotapi.NewMessage(chatID, emojiWarning+" unknown command. Please try again."))
		b.logger.Warnf("Failed to send message: %+v", err)
	}

}

func (b *Bot) handleCalendarCallback(callbackData string, chatID int64, state *UserState) {
	prefix, nums, err := ParseCalendarCallback(callbackData)
	if err != nil {
		err = b.Send(tgbotapi.NewMessage(chatID, emojiWarning+" "+err.Error()))
		b.logger.Warnf("Failed to send message: %+v", err)
		return
	}

	switch prefix {
	case CallbackSetYear:
		id := nums[0]
		year := nums[1]

		b.showMonthPicker(chatID, id, year)

	case CallbackSetMonth:
		id := nums[0]
		year := nums[1]
		month := nums[2]

		b.showDayPicker(chatID, id, year, month)

	case CallbackSetDay:
		id := nums[0]
		year := nums[1]
		month := nums[2]
		day := nums[3]

		debt, err := b.serviceDebt.GetDebt(b.ctx, id)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: "+err.Error()))
			b.logger.Warnf("Failed to send message: %+v", err)
			delete(state.TempData, "debt_id")
			delete(state.TempData, "return_date")
			state.ExpectedAction = ""
			return
		}

		date := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
		debt.ReturnDate = &date
		err = b.serviceDebt.UpdateDebt(b.ctx, debt)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" "+err.Error()))
			b.logger.Warnf("Failed to send message: %+v", err)
			delete(state.TempData, "debt_id")
			delete(state.TempData, "return_date")
			state.ExpectedAction = ""
			return
		}

		delete(state.TempData, "debt_id")
		delete(state.TempData, "return_date")
		state.ExpectedAction = ""

		b.showDebtMenu(chatID)
	default:
		b.logger.Warnf("Unknown callback: %s", prefix)

		err = b.Send(tgbotapi.NewMessage(chatID, emojiWarning+" unknown command. Please try again."))
		b.logger.Warnf("Failed to send message: %+v", err)
	}
}
