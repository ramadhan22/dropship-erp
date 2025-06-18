package service

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
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
	var total float64
	for _, l := range e.Lines {
		total += l.Amount
	}
	e.Amount = total
	if err := expRepo.Create(ctx, e); err != nil {
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
		return err
	}
	for i := range e.Lines {
		l := e.Lines[i]
		jl := &models.JournalLine{JournalID: jid, AccountID: l.AccountID, IsDebit: true, Amount: l.Amount, Memo: &e.Description}
		if err := jRepo.InsertJournalLine(ctx, jl); err != nil {
			return err
		}
	}
	jlAsset := &models.JournalLine{JournalID: jid, AccountID: e.AssetAccountID, IsDebit: false, Amount: total, Memo: &e.Description}
	if err := jRepo.InsertJournalLine(ctx, jlAsset); err != nil {
		return err
	}
	if tx != nil {
		return tx.Commit()
	}
	return nil
}

func (s *ExpenseService) ListExpenses(ctx context.Context, accountID int64, sortBy, dir string, limit, offset int) ([]models.Expense, int, error) {
	return s.expenseRepo.List(ctx, accountID, sortBy, dir, limit, offset)
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id string) error {
	return s.expenseRepo.Delete(ctx, id)
}
