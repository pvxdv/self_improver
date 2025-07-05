package telegram

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pvxdv/self_improver/internal/config/telegram"
	"github.com/pvxdv/self_improver/internal/service/finance/debt"
	"go.uber.org/zap"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	serviceDebt *debt.ServiceDebt
	logger      *zap.SugaredLogger

	ctx        context.Context
	password   string
	userStates map[int64]*UserState
}

type UserState struct {
	AuthPassed     bool
	ExpectedAction string // например: "add_debt_amount", "pay_debt_id"
	TempData       map[string]interface{}
}

func New(ctx context.Context, cfg *telegram.Config, logger *zap.SugaredLogger, serviceDebt *debt.ServiceDebt) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}
	botAPI.Debug = false

	return &Bot{
		api:         botAPI,
		serviceDebt: serviceDebt,
		logger:      logger,
		ctx:         ctx,
		password:    cfg.Password,
		userStates:  make(map[int64]*UserState),
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			go b.handleUpdate(update)
		case <-b.ctx.Done():
			return b.ctx.Err()
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	state, ok := b.userStates[chatID]
	if !ok {
		state = &UserState{TempData: make(map[string]interface{})}
		b.userStates[chatID] = state
	}

	if !state.AuthPassed {
		HandleAuth(b, chatID, update.Message.Text, state)
		return
	}

	switch update.Message.Text {
	case "Debts":
		b.showDebtMenu(chatID)
	case "Trends":
		b.showTrendPlaceholder(chatID)
	case "Back":
		b.showMainMenu(chatID)
	default:
		ProcessDebtActions(b, chatID, update.Message.Text, state)
	}
}

func (b *Bot) IsPassValid(pass string) bool {
	return b.password != pass
}

func (b *Bot) Send(msg tgbotapi.Chattable) error {
	_, err := b.api.Send(msg)
	return err
}
