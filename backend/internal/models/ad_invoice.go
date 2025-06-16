package models

import "time"

// AdInvoice represents an invoice for advertising spend.
type AdInvoice struct {
	InvoiceNo   string    `db:"invoice_no" json:"invoice_no"`
	Username    string    `db:"username" json:"username"`
	Store       string    `db:"store" json:"store"`
	InvoiceDate time.Time `db:"invoice_date" json:"invoice_date"`
	Total       float64   `db:"total" json:"total"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
