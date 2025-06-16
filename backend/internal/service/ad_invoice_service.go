package service

import (
	"bytes"
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	pdf "github.com/ledongthuc/pdf"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// AdInvoiceService handles ads invoice imports.
type AdInvoiceService struct {
	db          *sqlx.DB
	repo        *repository.AdInvoiceRepo
	journalRepo *repository.JournalRepo
}

func NewAdInvoiceService(db *sqlx.DB, r *repository.AdInvoiceRepo, jr *repository.JournalRepo) *AdInvoiceService {
	return &AdInvoiceService{db: db, repo: r, journalRepo: jr}
}

func formatStoreName(username string) string {
	u := strings.ToLower(strings.TrimSpace(username))
	if u == "" {
		return ""
	}
	if u == "mrest0re" {
		return "MR eStore Shopee"
	}
	u = strings.ReplaceAll(u, ".", " ")
	return strings.ToUpper(u[:1]) + u[1:]
}

func adsSaldoShopeeAccountID(store string) int64 {
	switch store {
	case "MR eStore Shopee":
		return 11011
	case "MR Barista Gear":
		return 11013
	default:
		return 11011
	}
}

func (s *AdInvoiceService) parsePDF(r io.Reader) (*models.AdInvoice, error) {
	tmp, err := os.CreateTemp("", "adinv*.pdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, r); err != nil {
		return nil, err
	}
	tmp.Close()

	reader, err := pdf.Open(tmp.Name())
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for i := 1; i <= reader.NumPage(); i++ {
		p := reader.Page(i)
		if p.V.IsNull() || p.V.Key("Contents").Kind() == pdf.Null {
			continue
		}
		buf.WriteString(p.GetPlainText("\n"))
	}
	txt := strings.Split(buf.String(), "\n")
	inv := &models.AdInvoice{}
	for i, line := range txt {
		line = strings.TrimSpace(line)
		switch line {
		case "No. Faktur":
			if i+1 < len(txt) {
				inv.InvoiceNo = strings.TrimSpace(txt[i+1])
			}
		case "Username":
			if i+2 < len(txt) {
				inv.Username = strings.TrimSpace(txt[i+2])
			}
		case "Tanggal Invoice":
			if i+1 < len(txt) {
				d := strings.TrimSpace(txt[i+1])
				inv.InvoiceDate, _ = time.Parse("02/01/2006", d)
			}
		}
		if strings.HasPrefix(line, "Total (") && i+1 < len(txt) {
			amt := strings.TrimSpace(txt[i+1])
			amt = strings.ReplaceAll(amt, ",", "")
			amt = strings.ReplaceAll(amt, ".", "")
			if v, err := strconv.ParseFloat(amt, 64); err == nil {
				inv.Total = v / 100
			}
		}
	}
	inv.Store = formatStoreName(inv.Username)
	return inv, nil
}

func (s *AdInvoiceService) ImportInvoicePDF(ctx context.Context, r io.Reader) error {
	inv, err := s.parsePDF(r)
	if err != nil {
		return err
	}
	inv.CreatedAt = time.Now()
	if err := s.repo.Insert(ctx, inv); err != nil {
		return err
	}
	if s.journalRepo != nil {
		je := &models.JournalEntry{
			EntryDate:    inv.InvoiceDate,
			Description:  strPtr("Shopee Ads " + inv.InvoiceNo),
			SourceType:   "ads_invoice",
			SourceID:     inv.InvoiceNo,
			ShopUsername: inv.Username,
			Store:        inv.Store,
			CreatedAt:    time.Now(),
		}
		jid, err := s.journalRepo.CreateJournalEntry(ctx, je)
		if err != nil {
			return err
		}
		jl1 := &models.JournalLine{JournalID: jid, AccountID: 52008, IsDebit: true, Amount: inv.Total, Memo: strPtr("Biaya Iklan " + inv.InvoiceNo)}
		jl2 := &models.JournalLine{JournalID: jid, AccountID: adsSaldoShopeeAccountID(inv.Store), IsDebit: false, Amount: inv.Total, Memo: strPtr("Pembayaran Iklan " + inv.InvoiceNo)}
		if err := s.journalRepo.InsertJournalLine(ctx, jl1); err != nil {
			return err
		}
		if err := s.journalRepo.InsertJournalLine(ctx, jl2); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdInvoiceService) ListInvoices(ctx context.Context, sortBy, dir string) ([]models.AdInvoice, error) {
	return s.repo.List(ctx, sortBy, dir)
}

func strPtr(s string) *string { return &s }
