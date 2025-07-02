package models

import "time"

type TaxPayment struct {
	ID          string    `db:"id" json:"id"`
	Store       string    `db:"store" json:"store"`
	PeriodType  string    `db:"period_type" json:"period_type"`
	PeriodValue string    `db:"period_value" json:"period_value"`
	Revenue     float64   `db:"revenue" json:"revenue"`
	TaxRate     float64   `db:"tax_rate" json:"tax_rate"`
	TaxAmount   float64   `db:"tax_amount" json:"tax_amount"`
	IsPaid      bool      `db:"is_paid" json:"is_paid"`
	PaidAt      time.Time `db:"paid_at" json:"paid_at"`
}
