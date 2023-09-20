package client

type Accrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

const (
	StatusRegistered = `REGISTERED`
	StatusInvalid    = `INVALID`
	StatusProcessing = `PROCESSING`
	StatusProcessed  = `PROCESSED`
)

const APIGetAccrual = `/api/orders/{number}`
