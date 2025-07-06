package telegram

import (
	"fmt"
	"strings"
	"time"
)

func formatAmount(amount int64) string {
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

func formatDate(date *time.Time) string {
	if date == nil {
		return "Not set"
	}
	return date.Format("2006-01-02")
}
