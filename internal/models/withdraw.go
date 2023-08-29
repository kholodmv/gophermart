package models

import "time"

type Withdraw struct {
	User        string     `json:"-"`
	Order       string     `json:"order"`
	Sum         int64      `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at"`
}
