package repository

import (
	"context"
	"fmt"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AdInvoiceRepo handles CRUD for ad_invoices table.
type AdInvoiceRepo struct{ db DBTX }

func NewAdInvoiceRepo(db DBTX) *AdInvoiceRepo { return &AdInvoiceRepo{db: db} }

func (r *AdInvoiceRepo) Insert(ctx context.Context, a *models.AdInvoice) error {
	_, err := r.db.NamedExecContext(ctx, `INSERT INTO ad_invoices (invoice_no, username, store, invoice_date, total)
                VALUES (:invoice_no,:username,:store,:invoice_date,:total)
                ON CONFLICT (invoice_no) DO NOTHING`, a)
	return err
}

func (r *AdInvoiceRepo) Exists(ctx context.Context, invoiceNo string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM ad_invoices WHERE invoice_no=$1)`, invoiceNo)
	return exists, err
}

func (r *AdInvoiceRepo) List(ctx context.Context, sortBy, dir string) ([]models.AdInvoice, error) {
	if sortBy == "" {
		sortBy = "invoice_date"
	}
	if dir != "asc" && dir != "desc" {
		dir = "desc"
	}
	query := fmt.Sprintf(`SELECT * FROM ad_invoices ORDER BY %s %s`, sortBy, dir)
	var res []models.AdInvoice
	err := r.db.SelectContext(ctx, &res, query)
	if res == nil {
		res = []models.AdInvoice{}
	}
	return res, err
}
