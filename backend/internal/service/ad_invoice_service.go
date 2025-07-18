package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// AdInvoiceService handles ads invoice imports.
type AdInvoiceService struct {
	db          *sqlx.DB
	repo        *repository.AdInvoiceRepo
	journalRepo *repository.JournalRepo
}

var amountRe = regexp.MustCompile(`-?[0-9][0-9.,]*`)

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

func parseAmount(s string) (float64, bool) {

	match := amountRe.FindString(s)
	if match == "" {
		return 0, false
	}
	clean := strings.ReplaceAll(match, ",", "")
	clean = strings.ReplaceAll(clean, ".", "")
	v, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, false
	}
	return v / 100, true
}

func parseInvoiceText(lines []string) *models.AdInvoice {
	inv := &models.AdInvoice{}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case line == "No. Faktur":
			var parts []string
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "" || next == "Username" || next == "Pelanggan" {
					if next == "Username" || next == "Pelanggan" {
						break
					}
					continue
				}
				if next == "Username" || strings.HasPrefix(next, "Username") {
					break
				}
				if next == "Periode" || strings.Contains(next, "Tanggal Invoice") {
					break
				}
				parts = append(parts, next)
			}
			inv.InvoiceNo = strings.Join(parts, "")
		case line == "Username":
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "" || next == "Pelanggan" {
					continue
				}
				inv.Username = next
				break
			}
		case strings.Contains(line, "Tanggal Invoice"):
			for j := i + 1; j < len(lines); j++ {
				d := strings.TrimSpace(lines[j])
				if t, err := time.Parse("02/01/2006", d); err == nil {
					inv.InvoiceDate = t
					break
				}
			}
		}

		if strings.HasPrefix(line, "Total") || strings.HasPrefix(line, "0.00") {
			cleaned := strings.TrimPrefix(line, "Total")
			cleaned = strings.TrimPrefix(cleaned, " (Termasuk PPN")
			cleaned = strings.TrimPrefix(cleaned, "0.00")
			if v, ok := parseAmount(cleaned); ok {
				inv.Total = v
				continue
			}
			var last float64
			for j := i + 1; j < len(lines); j++ {
				amt := strings.TrimSpace(lines[j])
				if amt == "" || strings.HasPrefix(amt, "(") {
					if last > 0 {
						break
					}
					continue
				}
				if v, ok := parseAmount(amt); ok {
					if v > 0 {
						last = v
					}
					continue
				}
				if last > 0 {
					break
				}
			}
			if last > 0 {
				inv.Total = last
			}
		}
	}
	inv.Store = formatStoreName(inv.Username)
	return inv
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

	if _, err := exec.LookPath("pdftotext"); err != nil {
		return nil, fmt.Errorf("pdftotext not installed: install poppler-utils")
	}
	out, err := exec.Command("pdftotext", tmp.Name(), "-").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	inv := parseInvoiceText(lines)
	return inv, nil
}

func (s *AdInvoiceService) ImportInvoicePDF(ctx context.Context, r io.Reader) error {
	inv, err := s.parsePDF(r)
	if err != nil {
		return err
	}
	if inv.Total <= 0 {
		return fmt.Errorf("failed to parse invoice: total <= 0")
	}
	inv.CreatedAt = time.Now()
	exists, err := s.repo.Exists(ctx, inv.InvoiceNo)
	if err != nil {
		return err
	}
	if exists {
		if err := s.repo.Delete(ctx, inv.InvoiceNo); err != nil {
			return err
		}
		if s.journalRepo != nil {
			old, err := s.journalRepo.GetJournalEntryBySource(ctx, "ads_invoice", inv.InvoiceNo)
			if err == nil && old != nil {
				if err := s.journalRepo.DeleteJournalEntry(ctx, old.JournalID); err != nil {
					return err
				}
			}
		}
	}
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
		lines := []models.JournalLine{
			{JournalID: jid, AccountID: 55003, IsDebit: true, Amount: inv.Total, Memo: strPtr("Biaya Iklan " + inv.InvoiceNo)},
			{JournalID: jid, AccountID: adsSaldoShopeeAccountID(inv.Store), IsDebit: false, Amount: inv.Total, Memo: strPtr("Pembayaran Iklan " + inv.InvoiceNo)},
		}
		// Use bulk insert for lines
		if err := s.journalRepo.InsertJournalLines(ctx, lines); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdInvoiceService) ListInvoices(ctx context.Context, sortBy, dir string) ([]models.AdInvoice, error) {
	return s.repo.List(ctx, sortBy, dir)
}

func strPtr(s string) *string { return &s }
