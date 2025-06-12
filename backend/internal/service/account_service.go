package service

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// AccountRepoInterface defines repo methods used by AccountService.
type AccountRepoInterface interface {
	CreateAccount(ctx context.Context, a *models.Account) (int64, error)
	GetAccountByID(ctx context.Context, id int64) (*models.Account, error)
	ListAccounts(ctx context.Context) ([]models.Account, error)
	UpdateAccount(ctx context.Context, a *models.Account) error
	DeleteAccount(ctx context.Context, id int64) error
}

// AccountService provides CRUD operations for accounts.
type AccountService struct {
	repo AccountRepoInterface
}

// NewAccountService constructs an AccountService.
func NewAccountService(r AccountRepoInterface) *AccountService {
	return &AccountService{repo: r}
}

func (s *AccountService) CreateAccount(ctx context.Context, a *models.Account) (int64, error) {
	return s.repo.CreateAccount(ctx, a)
}

func (s *AccountService) GetAccount(ctx context.Context, id int64) (*models.Account, error) {
	return s.repo.GetAccountByID(ctx, id)
}

func (s *AccountService) ListAccounts(ctx context.Context) ([]models.Account, error) {
	return s.repo.ListAccounts(ctx)
}

func (s *AccountService) UpdateAccount(ctx context.Context, a *models.Account) error {
	return s.repo.UpdateAccount(ctx, a)
}

func (s *AccountService) DeleteAccount(ctx context.Context, id int64) error {
	return s.repo.DeleteAccount(ctx, id)
}
