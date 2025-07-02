package service

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

type fakeAdjRepo struct {
	inserted []*models.ShopeeAdjustment
	deleted  []string
}

func (f *fakeAdjRepo) Insert(ctx context.Context, a *models.ShopeeAdjustment) error {
	cp := *a
	f.inserted = append(f.inserted, &cp)
	return nil
}

func (f *fakeAdjRepo) Delete(ctx context.Context, order string, t time.Time, typ string) error {
	f.deleted = append(f.deleted, order+"|"+t.Format("2006-01-02")+"|"+typ)
	return nil
}

func (f *fakeAdjRepo) List(ctx context.Context, from, to string) ([]models.ShopeeAdjustment, error) {
	return nil, nil
}

func (f *fakeAdjRepo) ListByOrder(ctx context.Context, order string) ([]models.ShopeeAdjustment, error) {
	return nil, nil
}

type fakeJournalRepoA struct {
	deleted []int64
	entries []*models.JournalEntry
}

func (f *fakeJournalRepoA) CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error) {
	e.JournalID = int64(len(f.entries) + 1)
	f.entries = append(f.entries, e)
	return e.JournalID, nil
}

func (f *fakeJournalRepoA) InsertJournalLine(ctx context.Context, l *models.JournalLine) error {
	return nil
}

func (f *fakeJournalRepoA) GetJournalEntryBySource(ctx context.Context, sourceType, sourceID string) (*models.JournalEntry, error) {
	for _, e := range f.entries {
		if e.SourceType == sourceType && e.SourceID == sourceID {
			return e, nil
		}
	}
	return nil, nil
}

func (f *fakeJournalRepoA) GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error) {
	return nil, nil
}

func (f *fakeJournalRepoA) UpdateJournalLineAmount(ctx context.Context, lineID int64, amount float64) error {
	return nil
}

func (f *fakeJournalRepoA) DeleteJournalEntry(ctx context.Context, id int64) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func TestShopeeAdjustmentImportDeletesExisting(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Adjustment")
	f.SetCellValue("Adjustment", "B2", "tokostore")
	f.SetCellValue("Adjustment", "A4", "Rincian Transaksi Penyesuaian")
	row := []interface{}{1, "2025-01-02", "Logistik", "missing", 100, "SO1"}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Adjustment", cell, v)
	}
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeAdjRepo{}
	jr := &fakeJournalRepoA{}
	svc := &ShopeeAdjustmentService{repo: repo, journalRepo: jr}
	inserted, err := svc.ImportXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil || inserted != 1 {
		t.Fatalf("import err %v inserted %d", err, inserted)
	}
	if len(repo.deleted) != 1 {
		t.Fatalf("delete not called")
	}
	if len(jr.deleted) != 0 {
		t.Fatalf("unexpected journal delete")
	}
	// import again to ensure deletion of previous journal
	inserted, err = svc.ImportXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil || inserted != 1 {
		t.Fatalf("second import err %v", err)
	}
	if len(jr.deleted) != 1 {
		t.Fatalf("journal not deleted on second import")
	}
}

func TestShopeeAdjustmentImportIgnoreBDMarketing(t *testing.T) {
	f := excelize.NewFile()
	sheet, _ := f.NewSheet("Adjustment")
	f.SetCellValue("Adjustment", "B2", "tokostore")
	f.SetCellValue("Adjustment", "A4", "Rincian Transaksi Penyesuaian")
	row := []interface{}{1, "2025-01-02", "BD Marketing", "fee", 100, "SO1"}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		f.SetCellValue("Adjustment", cell, v)
	}
	f.SetActiveSheet(sheet)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}

	repo := &fakeAdjRepo{}
	jr := &fakeJournalRepoA{}
	svc := &ShopeeAdjustmentService{repo: repo, journalRepo: jr}
	inserted, err := svc.ImportXLSX(context.Background(), bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import err %v", err)
	}
	if inserted != 0 {
		t.Fatalf("expected 0 rows inserted, got %d", inserted)
	}
	if len(repo.deleted) != 0 {
		t.Fatalf("unexpected delete calls")
	}
	if len(jr.entries) != 0 {
		t.Fatalf("unexpected journal entries")
	}
}
