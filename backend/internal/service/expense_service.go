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
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.expenseRepo.Create(ctx, e); err != nil {
		return err
	}
	je := &models.JournalEntry{
		EntryDate:    e.Date,
		Description:  &e.Description,
		SourceType:   "expense",
		SourceID:     e.ID,
		ShopUsername: "", // optional shop not tracked
		CreatedAt:    time.Now(),
	}
	jid, err := s.journalRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	jl1 := &models.JournalLine{JournalID: jid, AccountID: e.AccountID, IsDebit: true, Amount: e.Amount, Memo: &e.Description}
	jl2 := &models.JournalLine{JournalID: jid, AccountID: 1001, IsDebit: false, Amount: e.Amount, Memo: &e.Description}
	if err := s.journalRepo.InsertJournalLine(ctx, jl1); err != nil {
		return err
	}
	if err := s.journalRepo.InsertJournalLine(ctx, jl2); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *ExpenseService) ListExpenses(ctx context.Context) ([]models.Expense, error) {
	return s.expenseRepo.List(ctx)
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id string) error {
	return s.expenseRepo.Delete(ctx, id)
}
