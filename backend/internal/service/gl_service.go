package service

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type GLService struct{ repo *repository.JournalRepo }

func NewGLService(r *repository.JournalRepo) *GLService { return &GLService{repo: r} }

func (s *GLService) FetchGeneralLedger(ctx context.Context, shop string, from, to time.Time) ([]repository.AccountBalance, error) {
	return s.repo.GetAccountBalancesBetween(ctx, shop, from, to)
}
