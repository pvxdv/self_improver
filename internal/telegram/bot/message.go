package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pvxdv/self_improver/internal/model"
	"strconv"
	"strings"
	"time"
)

func (b *Bot) handleIncomingMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	state := b.getUserState(chatID)

	b.logger.Debug("Incoming message. ChatID: %d, Text: %s, State: %+v", chatID, msg.Text, state)

	if msg.IsCommand() && msg.Command() == "start" {
		state.AuthPassed = false
		state.ExpectedAction = ""
		state.TempData = make(map[string]interface{})
		b.showAuthMenu(chatID)
		return
	}

	if !state.AuthPassed {
		b.handleAuth(chatID, msg.Text, state)
		return
	}

	b.logger.Debugf("Expected action: %s", state.ExpectedAction)

	switch state.ExpectedAction {
	case "add_debt_amount":
		amount, err := strconv.ParseInt(msg.Text, 10, 64)
		if err != nil || amount <= 0 {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" please enter a valid positive amount:"))
			b.logger.Warnf("Failed to send message: %+v", err)
			return
		}

		b.logger.Debugf("Received amount: %d", amount)

		state.TempData["amount"] = amount
		state.ExpectedAction = "add_debt_description"

		err = b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" now please enter debt description:"))
		b.logger.Warnf("Failed to send message: %+v", err)

	case "add_debt_description":
		amount, ok := state.TempData["amount"].(int64)
		if !ok {
			b.logger.Error("amount not found in TempData")
			err := b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: amount not found. Please start over with /add"))
			b.logger.Warnf("Failed to send message: %+v", err)
			state.ExpectedAction = ""
			return
		}

		desc := strings.TrimSpace(msg.Text)
		if len(desc) < 2 {
			err := b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" Description too short. Minimum 2 characters required:"))
			b.logger.Warnf("Failed to send message: %+v", err)
			return
		}

		b.logger.Debugf("Creating new debt. Amount: %d, Description: %s", amount, desc)

		state.TempData["description"] = desc
		state.ExpectedAction = "add_debt_return_date"

		msgText := fmt.Sprintf("%s Enter due date (YYYY-MM-DD), or send /skip if not needed:", emojiCalendar)
		err := b.Send(tgbotapi.NewMessage(chatID, msgText))
		b.logger.Warnf("Failed to send message: %+v", err)

	case "add_debt_return_date":
		desc, ok := state.TempData["description"].(string)
		amount, ok2 := state.TempData["amount"].(int64)
		if !ok || !ok2 {
			err := b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: missing data. Please try again."))
			b.logger.Warnf("Failed to send message: %+v", err)
			state.ExpectedAction = ""
			return
		}

		var dueDate *time.Time
		if msg.Text != "/skip" {
			parsedDate, err := time.Parse("2006-01-02", msg.Text)
			if err != nil {
				err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" invalid date format. Use YYYY-MM-DD:"))
				b.logger.Warnf("Failed to send message: %+v", err)
				return
			}
			dueDate = &parsedDate
		}

		err := b.serviceDebt.AddDebt(b.ctx, &model.Debt{
			Description: desc,
			Amount:      amount,
			ReturnDate:  dueDate,
		})
		if err != nil {
			b.logger.Errorf("failed to add debt: %v", err)
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" failed to save debt."))
			b.logger.Warnf("Failed to send message: %+v", err)
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
				emojiSuccess+" Debt added successfully!\n\n"+
					emojiMoney+": %d\n"+
					emojiInfo+": %s\n"+
					emojiCalendar+": %s",
				amount, desc, b.formatDate(dueDate),
			)))
			b.logger.Warnf("Failed to send message: %+v", err)
		}

		state.ExpectedAction = ""
		delete(state.TempData, "amount")
		delete(state.TempData, "description")

		time.Sleep(500 * time.Millisecond)
		b.showDebtMenu(chatID)

	case "pay_debt_amount":
		amount, err := strconv.ParseInt(msg.Text, 10, 64)
		if err != nil || amount <= 0 {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" please enter a valid positive payment amount:"))
			b.logger.Warnf("Failed to send message: %+v", err)
			return
		}

		debtID, ok := state.TempData["debt_id"].(int64)
		if !ok {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: debt not found. Please try again."))
			b.logger.Warnf("Failed to send message: %+v", err)
			state.ExpectedAction = ""
			return
		}

		err = b.serviceDebt.PayDebt(b.ctx, debtID, amount)
		if err != nil {
			b.logger.Errorf("Failed to pay debt: %v", err)
			err = b.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(emojiFailed+" payment failed: %v", err)))
			b.logger.Warnf("Failed to send message: %+v", err)
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiSuccess+" payment successful!"))
			b.logger.Warnf("Failed to send message: %+v", err)
		}

		state.ExpectedAction = ""
		delete(state.TempData, "debt_id")

		b.showDebtMenu(chatID)

	case "update_debt_amount":
		debtID, ok := state.TempData["debt_id"].(int64)
		if !ok {
			b.logger.Error("DebtID not found in TempData")

			err := b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: debt not found. Please try again."))
			b.logger.Warnf("Failed to send message: %+v", err)

			state.ExpectedAction = ""
			return
		}

		amount, err := strconv.ParseInt(msg.Text, 10, 64)
		if err != nil || amount <= 0 {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" please enter a valid positive amount:"))
			b.logger.Warnf("Failed to send message: %+v", err)
			return
		}

		b.logger.Infof("Updating debt #%d amount to %d", debtID, amount)

		debt, err := b.serviceDebt.GetDebt(b.ctx, debtID)
		if err != nil {
			b.logger.Errorf("Failed to get debt: %v", err)

			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: debt not found."))
			b.logger.Warnf("Failed to send message: %+v", err)

			state.ExpectedAction = ""
			return
		}

		debt.Amount = amount
		if err = b.serviceDebt.UpdateDebt(b.ctx, debt); err != nil {
			b.logger.Errorf("Failed to update debt: %v", err)

			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" failed to update amount."))
			b.logger.Warnf("Failed to send message: %+v", err)
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiSuccess+" amount updated successfully!"))
			b.logger.Warnf("Failed to send message: %+v", err)
		}

		state.ExpectedAction = ""
		delete(state.TempData, "debt_id")

		time.Sleep(300 * time.Millisecond)

		b.showDebtMenu(chatID)

	case "update_debt_description":
		debtID, ok := state.TempData["debt_id"].(int64)
		if !ok {
			b.logger.Error("debtID not found in TempData")
			err := b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: debt not found. Please try again."))
			b.logger.Warnf("Failed to send message: %+v", err)
			state.ExpectedAction = ""
			return
		}

		desc := strings.TrimSpace(msg.Text)
		if len(desc) < 2 {
			err := b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" description too short. Minimum 2 characters required:"))
			b.logger.Warnf("Failed to send message: %+v", err)
			return
		}

		b.logger.Debugf("updating debt #%d description to '%s'", debtID, desc)

		debt, err := b.serviceDebt.GetDebt(b.ctx, debtID)
		if err != nil {
			b.logger.Errorf("failed to get debt: %v", err)
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: debt not found."))
			b.logger.Warnf("Failed to send message: %+v", err)
			state.ExpectedAction = ""
			return
		}

		debt.Description = desc
		err = b.serviceDebt.UpdateDebt(b.ctx, debt)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" failed to update description."))
			b.logger.Warnf("Failed to send message: %+v", err)
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiSuccess+" description updated successfully!"))
			b.logger.Warnf("Failed to send message: %+v", err)
		}

		state.ExpectedAction = ""
		delete(state.TempData, "debt_id")
		delete(state.TempData, "updated_debt")

		time.Sleep(500 * time.Millisecond)
		b.showDebtMenu(chatID)

	case "update_debt_return_date":
		debtID, ok := state.TempData["debt_id"].(int64)
		if !ok {
			err := b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: debt not found. Please try again."))
			b.logger.Warnf("Failed to send message: %+v", err)
			state.ExpectedAction = ""
			return
		}

		var returnDate *time.Time
		if msg.Text != "/skip" {
			parsedDate, err := time.Parse("2006-01-02", msg.Text)
			if err != nil {
				err = b.Send(tgbotapi.NewMessage(chatID, emojiCalendar+" invalid date format. Use YYYY-MM-DD:"))
				b.logger.Warnf("Failed to send message: %+v", err)
				return
			}
			returnDate = &parsedDate
		}

		debt, err := b.serviceDebt.GetDebt(b.ctx, debtID)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" error: debt not found."))
			b.logger.Warnf("Failed to send message: %+v", err)
			state.ExpectedAction = ""
			return
		}

		debt.ReturnDate = returnDate
		err = b.serviceDebt.UpdateDebt(b.ctx, debt)
		if err != nil {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiFailed+" failed to update return date."))
			b.logger.Warnf("Failed to send message: %+v", err)
		} else {
			err = b.Send(tgbotapi.NewMessage(chatID, emojiSuccess+" return date updated successfully!"))
			b.logger.Warnf("Failed to send message: %+v", err)
		}

		state.ExpectedAction = ""
		delete(state.TempData, "debt_id")

		time.Sleep(500 * time.Millisecond)
		b.showDebtMenu(chatID)

	default:
		b.logger.Warnf("Unexpected message in state: %s", state.ExpectedAction)

		err := b.Send(tgbotapi.NewMessage(chatID, emojiInfo+" please use the menu buttons to interact with the bot."))
		b.logger.Warnf("Failed to send message: %+v", err)
		b.showMainMenu(chatID)
	}
}
