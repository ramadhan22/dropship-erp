package models

import "time"

type Expense struct {
	ID             string        `db:"id" json:"id"`
	Date           time.Time     `db:"date" json:"date"`
	Description    string        `db:"description" json:"description"`
	Amount         float64       `db:"amount" json:"amount"`
	AssetAccountID int64         `db:"asset_account_id" json:"asset_account_id"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	Lines          []ExpenseLine `json:"lines"`
}

type ExpenseLine struct {
	LineID    int64   `db:"line_id" json:"line_id"`
	ExpenseID string  `db:"expense_id" json:"expense_id"`
	AccountID int64   `db:"account_id" json:"account_id"`
	Amount    float64 `db:"amount" json:"amount"`
}
