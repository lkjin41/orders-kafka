package orders

import "time"

type Order struct {
	ID          int64     `json:"id"`
	CustomerID  string    `json:"customer_id"`
	Status      string    `json:"status"`
	TotalAmount int       `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	CustomerID  string `json:"customer_id"`
	TotalAmount int    `json:"total_amount"`
}
