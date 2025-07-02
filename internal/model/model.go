package model

import "time"

type Trend struct {
	ID         int64        `json:"id,omitempty"`
	Name       string       `json:"name"`
	Challenges []*Challenge `json:"challenges,omitempty"`
}

type Challenge struct {
	ID          int64      `json:"id,omitempty"`
	TrendID     string     `json:"trend_id,omitempty"`
	TrendName   string     `json:"trend_name"`
	Description string     `json:"description"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}
