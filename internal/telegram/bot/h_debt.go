package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pvxdv/self_improver/internal/model"
	"strconv"
	"strings"
)

func (b *Bot) showDebtListAndMenu(chatID int64) {
	debts, err := b.serviceDebt.GetAmountDebt(b.ctx)
	if err != nil || len(debts.Details) == 0 {
		msg := tgbotapi.NewMessage(chatID, "No debts found.")
		b.Send(msg)
		b.showDebtMenu(chatID)
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total debt amount: %s\n\n", formatAmount(debts.Amount)))
	sb.WriteString("Your debts:\n")

	for _, d := range debts.Details {
		sb.WriteString(fmt.Sprintf(
			"ID: %d — %s: %s\n",
			d.ID,
			d.Description,
			formatAmount(d.Amount),
		))
	}

	msg := tgbotapi.NewMessage(chatID, sb.String())
	if err := b.Send(msg); err != nil {
		b.logger.Errorw("Failed to send debt list", "error", err)
	}

	b.showDebtMenu(chatID)
}

func formatAmount(amount int64) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}

	s := strconv.FormatInt(amount, 10)
	n := len(s)

	var result strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteByte(s[i])
	}

	if neg {
		return "- " + result.String() + " ₽"
	}
	return result.String() + " ₽"
}

func ProcessDebtActions(b *Bot, chatID int64, text string, state *UserState) {
	switch state.ExpectedAction {
	case "add_debt_amount":
		amount, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, "Invalid amount. Please enter a number."))
			if err != nil {
				b.logger.Error(err)
			}
			return
		}
		state.TempData["amount"] = amount
		state.ExpectedAction = "add_debt_description"
		err = b.Send(tgbotapi.NewMessage(chatID, "Enter debt description:"))
		if err != nil {
			b.logger.Errorw("Failed to send message", "error", err)
		}

	case "add_debt_description":
		desc := text
		amount := state.TempData["amount"].(int64)
		debt := &model.Debt{
			Description: desc,
			Amount:      amount,
		}
		err := b.serviceDebt.AddDebt(b.ctx, debt)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Failed to add debt: %v", err)))
			if err != nil {
				b.logger.Error(err)
			}
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, "Debt added successfully!"))
			if err != nil {
				b.logger.Error(err)
			}
		}
		state.ExpectedAction = ""
		b.showDebtListAndMenu(chatID)

	case "pay_debt_id":
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, "Invalid ID. Please enter a valid debt ID."))
			if err != nil {
				b.logger.Error(err)
			}
			return
		}
		state.TempData["debt_id"] = id
		state.ExpectedAction = "pay_debt_amount"
		err = b.Send(tgbotapi.NewMessage(chatID, "Enter payment amount:"))
		if err != nil {
			b.logger.Errorw("Failed to send message", "error", err)
		}

	case "pay_debt_amount":
		amount, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, "Invalid amount. Please enter a number."))
			if err != nil {
				b.logger.Error(err)
			}
			return
		}
		id := state.TempData["debt_id"].(int64)
		err = b.serviceDebt.PayDebt(b.ctx, id, amount)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Payment error: %v", err)))
			if err != nil {
				b.logger.Error(err)
			}
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, "Payment successful!"))
			if err != nil {
				b.logger.Error(err)
			}
		}
		state.ExpectedAction = ""
		b.showDebtListAndMenu(chatID)

	case "delete_debt_id":
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, "Invalid ID. Please enter a valid debt ID."))
			if err != nil {
				b.logger.Error(err)
			}
			return
		}
		err = b.serviceDebt.DeleteDebt(b.ctx, id)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Failed to delete debt: %v", err)))
			if err != nil {
				b.logger.Error(err)
			}
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, "Debt deleted successfully!"))
			if err != nil {
				b.logger.Error(err)
			}
		}
		state.ExpectedAction = ""
		b.showDebtListAndMenu(chatID)

	case "update_debt_id":
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, "Invalid ID. Please enter a valid debt ID."))
			if err != nil {
				b.logger.Error(err)
			}
			return
		}

		debt, err := b.serviceDebt.GetDebt(b.ctx, id)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, "Failed to get debt for update."))
			if err != nil {
				b.logger.Error(err)
			}
			return
		}

		state.TempData["debt"] = debt
		state.ExpectedAction = "update_debt_choice"

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Update Amount"),
				tgbotapi.NewKeyboardButton("Update Description"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Cancel Update"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, "What do you want to update?")
		msg.ReplyMarkup = keyboard
		err = b.Send(msg)
		if err != nil {
			b.logger.Error(err)
		}

	case "update_debt_choice":
		_ = state.TempData["debt"].(model.Debt)

		switch text {
		case "Update Amount":
			state.ExpectedAction = "update_debt_amount"
			err := b.Send(tgbotapi.NewMessage(chatID, "Enter new amount:"))
			if err != nil {
				b.logger.Errorw("Failed to send message", "error", err)
			}

		case "Update Description":
			state.ExpectedAction = "update_debt_description"
			err := b.Send(tgbotapi.NewMessage(chatID, "Enter new description:"))
			if err != nil {
				b.logger.Errorw("Failed to send message", "error", err)
			}

		case "Cancel Update":
			state.ExpectedAction = ""
			delete(state.TempData, "debt")
			b.showDebtListAndMenu(chatID)

		default:
			err := b.Send(tgbotapi.NewMessage(chatID, "Please choose an option from the menu."))
			if err != nil {
				b.logger.Error(err)
			}
		}
	case "update_debt_amount":
		switch text {
		case "Update Amount", "Update Description", "Cancel Update":
			if text == "Cancel Update" {
				state.ExpectedAction = ""
				delete(state.TempData, "debt")
				b.Send(tgbotapi.NewMessage(chatID, "Update cancelled."))
				b.showDebtListAndMenu(chatID)
			} else {
				b.Send(tgbotapi.NewMessage(chatID, "Please finish entering the amount first."))
			}
			return
		}

		amount, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			b.Send(tgbotapi.NewMessage(chatID, "Invalid amount. Please enter a number."))
			return
		}

		debt := state.TempData["debt"].(model.Debt)
		debt.Amount = amount

		err = b.serviceDebt.UpdateDebt(b.ctx, debt)
		if err != nil {
			b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Failed to update debt: %v", err)))
		} else {
			b.Send(tgbotapi.NewMessage(chatID, "Amount updated successfully!"))
		}
		state.ExpectedAction = ""
		delete(state.TempData, "debt")
		b.showDebtListAndMenu(chatID)
	case "update_debt_description":
		switch text {
		case "Update Amount", "Update Description", "Cancel Update":
			if text == "Cancel Update" {
				state.ExpectedAction = ""
				delete(state.TempData, "debt")
				b.Send(tgbotapi.NewMessage(chatID, "Update cancelled."))
				b.showDebtListAndMenu(chatID)
			} else {
				b.Send(tgbotapi.NewMessage(chatID, "Please finish entering the description first."))
			}
			return
		}

		// Если всё ок — обновляем описание
		debt := state.TempData["debt"].(model.Debt)
		debt.Description = text

		err := b.serviceDebt.UpdateDebt(b.ctx, debt)
		if err != nil {
			b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Failed to update debt: %v", err)))
		} else {
			b.Send(tgbotapi.NewMessage(chatID, "Description updated successfully!"))
		}
		state.ExpectedAction = ""
		delete(state.TempData, "debt")
		b.showDebtListAndMenu(chatID)
	}

	switch text {
	case "Add Debt":
		state.ExpectedAction = "add_debt_amount"
		err := b.Send(tgbotapi.NewMessage(chatID, "Enter debt amount:"))
		if err != nil {
			b.logger.Errorw("Failed to send message", "error", err)
		}

	case "List Debts":
		b.showDebtListAndMenu(chatID)

	case "Pay Debt":
		state.ExpectedAction = "pay_debt_id"
		err := b.Send(tgbotapi.NewMessage(chatID, "Enter debt ID:"))
		if err != nil {
			b.logger.Errorw("Failed to send message", "error", err)
		}

	case "Delete Debt":
		state.ExpectedAction = "delete_debt_id"
		err := b.Send(tgbotapi.NewMessage(chatID, "Enter debt ID to delete:"))
		if err != nil {
			b.logger.Errorw("Failed to send message", "error", err)
		}

	case "Update Debt":
		state.ExpectedAction = "update_debt_id"
		err := b.Send(tgbotapi.NewMessage(chatID, "Enter debt ID to update:"))
		if err != nil {
			b.logger.Errorw("Failed to send message", "error", err)
		}
	}
}
