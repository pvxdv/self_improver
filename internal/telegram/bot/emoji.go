package telegram

const (
	emojiSuccess  = "✅"
	emojiFailed   = "❌"
	emojiInfo     = "ℹ️"
	emojiMoney    = "💰"
	emojiId       = "📌"
	emojiAuth     = "🔐"
	emojiTrends   = "📈"
	emojiDebts    = "💸"
	emojiWarning  = "⚠️"
	emojiPoint    = "🔹"
	emojiCalendar = "📅"
	emojiDeadline = ""

	emojiPointGreen  = "🟢"
	emojiPointYellow = "🟡"
	emojiPointOrange = "🟠"
	emojiPointRed    = "🔴"
	emojiPointBlack  = "⚫"

	emojiActionAdd    = "🆕"
	emojiActionBack   = "⬅"
	emojiActionPay    = "💳"
	emojiActionEdit   = "⚙️"
	emojiActionDelete = "🗑"
)

func (b *Bot) getDeadlineIcon(daysLeft int) string {
	switch {
	case daysLeft > 31:
		return emojiPointGreen
	case daysLeft > 14:
		return emojiPointYellow
	case daysLeft > 7:
		return emojiPointOrange
	case daysLeft > 0:
		return emojiPointRed
	default:
		return emojiPointBlack
	}
}
