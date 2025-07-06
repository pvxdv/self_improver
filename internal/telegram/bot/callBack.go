package telegram

import (
	"fmt"
	"strconv"
	"strings"

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
)

func CallbackWithData(prefix string, id int64) string {
	return fmt.Sprintf("%s:%d", prefix, id)
}

func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	callbackData := callback.Data

	if _, err := b.api.Request(tgbotapi.NewCallback(callback.ID, "")); err != nil {
		b.logger.Warnw("Failed to answer callback query", "error", err)
	}

	state := b.getUserState(chatID)
	b.logger.Debug("Processing callback: %s", callbackData)

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

		b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" enter debt amount:"))
		return
	case CallbackAuthorize:
		state.ExpectedAction = "await_password"

		b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" enter your password:"))
		return
	case CallbackTrends:
		b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" feature is not implemented yet"))
		return
	case CallbackCancelDelete:
		b.Send(tgbotapi.NewMessage(chatID, emojiActionDelete+" deletion cancelled"))

		b.showDebtMenu(chatID)
		return
	default:

	}

	parts := strings.SplitN(callbackData, ":", 2)
	if len(parts) != 2 {
		b.logger.Warnf("Invalid callback format: %s", callbackData)

		b.Send(tgbotapi.NewMessage(chatID, emojiWarning+" invalid command format. Please try again."))
		return
	}

	prefix := parts[0]
	idStr := parts[1]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		b.logger.Warnf("Invalid ID in callback: %s", idStr)

		b.Send(tgbotapi.NewMessage(chatID, emojiWarning+" invalid ID format. Please try again."))
		return
	}

	switch prefix {
	case CallbackUpdateReturnDate:
		state.ExpectedAction = "update_debt_return_date"
		state.TempData["debt_id"] = id

		debt, err := b.serviceDebt.GetDebt(b.ctx, id)
		if err != nil {
			b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" Debt not found."))
			return
		}

		currentDate := "Not set"
		if debt.ReturnDate != nil {
			currentDate = debt.ReturnDate.Format("2006-01-02")
		}

		msgText := fmt.Sprintf(
			"%s Current return date: %s\nEnter new date (YYYY-MM-DD), or send /skip to remove it:",
			emojiCalendar, currentDate,
		)
		b.Send(tgbotapi.NewMessage(chatID, msgText))

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
		b.Send(msg)

	case CallbackUpdateAmount:
		state.ExpectedAction = "update_debt_amount"
		state.TempData["debt_id"] = id

		b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiInfo+" enter new amount for Debt #%d:", id)))

	case CallbackUpdateDescription:
		state.ExpectedAction = "update_debt_description"
		state.TempData["debt_id"] = id

		b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiInfo+" enter new description for Debt #%d:", id)))

	case CallbackPayDebt:
		state.ExpectedAction = "pay_debt_amount"
		state.TempData["debt_id"] = id

		b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" enter payment amount:"))

	case CallbackDeleteDebt:
		debt, err := b.serviceDebt.GetDebt(b.ctx, id)
		if err != nil {
			b.Send(tgbotapi.NewMessage(chatID, emojiFailed+"️ error: Debt not found"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			emojiWarning+"️ Confirm Deletion\n\n"+
				"You're about to delete:\n\n"+
				emojiPoint+" "+emojiId+": %d\n"+
				emojiPoint+" "+emojiInfo+": %s\n"+
				emojiPoint+" "+emojiMoney+": %s\n\n"+
				"This cannot be undone!",
			id, debt.Description, formatAmount(debt.Amount)))

		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(emojiActionDelete+" delete", fmt.Sprintf("%s:%d", CallbackConfirmDelete, id)),
				tgbotapi.NewInlineKeyboardButtonData(emojiActionBack+" back", CallbackCancelDelete),
			),
		)

		b.Send(msg)

	case CallbackConfirmDelete:
		if err := b.serviceDebt.DeleteDebt(b.ctx, id); err != nil {
			b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" failed to delete debt: "+err.Error()))
		} else {
			b.Send(tgbotapi.NewMessage(chatID, emojiSuccess+" debt deleted successfully"))
		}

		b.showDebtMenu(chatID)

	default:
		b.logger.Warnf("Unknown callback prefix: %s", prefix)

		b.Send(tgbotapi.NewMessage(chatID, emojiWarning+" unknown command. Please try again."))
	}
}
