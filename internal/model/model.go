package model

import "time"

// Trend
// @Description Trend is a personal improvement goal.
// @Description It can contain multiple challenges.
type Trend struct {
	ID         int64        `json:"id,omitempty" example:"1"`
	Name       string       `json:"name" example:"programming"`
	Challenges []*Challenge `json:"challenges,omitempty"`
}

// Challenge
// @Description Challenge is a specific task within a trend.
type Challenge struct {
	ID          int64      `json:"id,omitempty" example:"1"`
	TrendID     int64      `json:"trend_id,omitempty" example:"1"`
	TrendName   string     `json:"trend_name" example:"programming"`
	Description string     `json:"description" example:"practice LeetCode for an hour a day"`
	StartDate   *time.Time `json:"start_date,omitempty" example:"2025-01-02T15:04:05Z"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"2025-01-02T15:04:05Z"`
}

// Debt
// @Description Represents a debt position, such as a credit or loan to friends/family.
// @Description Contains basic information about the debt obligation.
type Debt struct {
	ID          int64  `json:"id,omitempty" example:"1"`
	Description string `json:"description" example:"Loan for car purchase"`
	Amount      int64  `json:"amount" example:"1000000"`
}

// AmountDebt
// @Description Represents aggregated debt information including:
// @Description - Total sum of all debts
// @Description - Detailed breakdown of individual debt positions
type AmountDebt struct {
	Amount  int64   `json:"amount" example:"1000000"`
	Details []*Debt `json:"details"`
}
