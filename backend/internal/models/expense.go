package models

import "time"

type Expense struct {
	ID          string    `db:"id" json:"id"`
	Date        time.Time `db:"date" json:"date"`
	Description string    `db:"description" json:"description"`
	Amount      float64   `db:"amount" json:"amount"`
	AccountID   int64     `db:"account_id" json:"account_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
