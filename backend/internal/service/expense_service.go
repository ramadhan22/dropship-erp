package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type ExpenseService struct {
	db          *sqlx.DB
	expenseRepo *repository.ExpenseRepo
	journalRepo *repository.JournalRepo
}

func NewExpenseService(db *sqlx.DB, er *repository.ExpenseRepo, jr *repository.JournalRepo) *ExpenseService {
	return &ExpenseService{db: db, expenseRepo: er, journalRepo: jr}
}

func (s *ExpenseService) CreateExpense(ctx context.Context, e *models.Expense) error {
	log.Printf("CreateExpense %s", e.ID)
	var tx *sqlx.Tx
	expRepo := s.expenseRepo
	jRepo := s.journalRepo
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	if s.db != nil {
		var err error
		tx, err = s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		expRepo = repository.NewExpenseRepo(tx)
		jRepo = repository.NewJournalRepo(tx)
	}
	var total float64
	for _, l := range e.Lines {
		total += l.Amount
	}
	e.Amount = total
	if err := expRepo.Create(ctx, e); err != nil {
		logutil.Errorf("CreateExpense error: %v", err)
		return err
	}
	je := &models.JournalEntry{
		EntryDate:    e.Date,
		Description:  &e.Description,
		SourceType:   "expense",
		SourceID:     e.ID,
		ShopUsername: "", // optional shop not tracked
		Store:        "",
		CreatedAt:    time.Now(),
	}
	jid, err := jRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		logutil.Errorf("CreateExpense journal error: %v", err)
		return err
	}
	// Prepare all journal lines for bulk insert
	lines := make([]models.JournalLine, 0, len(e.Lines)+1)
	for i := range e.Lines {
		l := e.Lines[i]
		lines = append(lines, models.JournalLine{
			JournalID: jid, 
			AccountID: l.AccountID, 
			IsDebit:   true, 
			Amount:    l.Amount, 
			Memo:      &e.Description,
		})
	}
	// Add asset line
	lines = append(lines, models.JournalLine{
		JournalID: jid, 
		AccountID: e.AssetAccountID, 
		IsDebit:   false, 
		Amount:    total, 
		Memo:      &e.Description,
	})
	
	// Use bulk insert for all lines
	if err := jRepo.InsertJournalLines(ctx, lines); err != nil {
		logutil.Errorf("CreateExpense lines error: %v", err)
		return err
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("CreateExpense committed %s", e.ID)
		return nil
	}
	log.Printf("CreateExpense done %s", e.ID)
	return nil
}

func (s *ExpenseService) ListExpenses(ctx context.Context, accountID int64, sortBy, dir string, limit, offset int) ([]models.Expense, int, error) {
	return s.expenseRepo.List(ctx, accountID, sortBy, dir, limit, offset)
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id string) error {
	log.Printf("DeleteExpense %s", id)
	err := s.expenseRepo.Delete(ctx, id)
	if err != nil {
		logutil.Errorf("DeleteExpense error: %v", err)
	}
	return err
}

func (s *ExpenseService) GetExpense(ctx context.Context, id string) (*models.Expense, error) {
	return s.expenseRepo.GetByID(ctx, id)
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, e *models.Expense) error {
	log.Printf("UpdateExpense %s", e.ID)
	var tx *sqlx.Tx
	expRepo := s.expenseRepo
	jRepo := s.journalRepo
	if s.db != nil {
		var err error
		tx, err = s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		expRepo = repository.NewExpenseRepo(tx)
		jRepo = repository.NewJournalRepo(tx)
	}

	oldEntry, err := jRepo.GetJournalEntryBySource(ctx, "expense", e.ID)
	if err == nil && oldEntry != nil {
		lines, _ := jRepo.GetLinesByJournalID(ctx, oldEntry.JournalID)
		rev := &models.JournalEntry{
			EntryDate:    time.Now(),
			Description:  expPtrString("Reverse " + e.Description),
			SourceType:   "expense_reverse",
			SourceID:     e.ID + "-rev-" + time.Now().Format("20060102150405"),
			ShopUsername: "",
			Store:        "",
			CreatedAt:    time.Now(),
		}
		jid, err := jRepo.CreateJournalEntry(ctx, rev)
		if err != nil {
			return err
		}
		// Prepare reverse lines for bulk insert
		reverseLines := make([]models.JournalLine, 0, len(lines))
		for _, l := range lines {
			reverseLines = append(reverseLines, models.JournalLine{
				JournalID: jid, 
				AccountID: l.AccountID, 
				IsDebit:   !l.IsDebit, 
				Amount:    l.Amount, 
				Memo:      rev.Description,
			})
		}
		// Use bulk insert for reverse lines
		if err := jRepo.InsertJournalLines(ctx, reverseLines); err != nil {
			return err
		}
	}

	var total float64
	for _, l := range e.Lines {
		total += l.Amount
	}
	e.Amount = total
	if err := expRepo.Update(ctx, e); err != nil {
		logutil.Errorf("UpdateExpense repo error: %v", err)
		return err
	}
	je := &models.JournalEntry{
		EntryDate:    e.Date,
		Description:  &e.Description,
		SourceType:   "expense",
		SourceID:     e.ID,
		ShopUsername: "",
		Store:        "",
		CreatedAt:    time.Now(),
	}
	jid, err := jRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		logutil.Errorf("UpdateExpense journal error: %v", err)
		return err
	}
	// Prepare all journal lines for bulk insert
	lines := make([]models.JournalLine, 0, len(e.Lines)+1)
	for i := range e.Lines {
		l := e.Lines[i]
		lines = append(lines, models.JournalLine{
			JournalID: jid, 
			AccountID: l.AccountID, 
			IsDebit:   true, 
			Amount:    l.Amount, 
			Memo:      &e.Description,
		})
	}
	// Add asset line
	lines = append(lines, models.JournalLine{
		JournalID: jid, 
		AccountID: e.AssetAccountID, 
		IsDebit:   false, 
		Amount:    total, 
		Memo:      &e.Description,
	})
	
	// Use bulk insert for all lines
	if err := jRepo.InsertJournalLines(ctx, lines); err != nil {
		logutil.Errorf("UpdateExpense lines error: %v", err)
		return err
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("UpdateExpense committed %s", e.ID)
		return nil
	}
	log.Printf("UpdateExpense done %s", e.ID)
	return nil
}

func expPtrString(s string) *string { return &s }
