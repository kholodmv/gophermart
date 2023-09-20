package withdraw

import "time"

type Withdraw struct {
	User        string     `json:"-"`
	Order       string     `json:"order"`
	Sum         float32    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at"`
}
