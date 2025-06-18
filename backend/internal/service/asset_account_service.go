package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

type AssetAccountRepo interface {
	Create(ctx context.Context, a *models.AssetAccount) error
	GetByID(ctx context.Context, id int64) (*models.AssetAccount, error)
	List(ctx context.Context) ([]models.AssetAccount, error)
}

type AssetAccountService struct {
	repo        AssetAccountRepo
	journalRepo *repository.JournalRepo
}

type AssetAccountBalance struct {
	AssetID   int64   `json:"asset_id"`
	AccountID int64   `json:"account_id"`
	Balance   float64 `json:"balance"`
}

func NewAssetAccountService(r AssetAccountRepo, jr *repository.JournalRepo) *AssetAccountService {
	return &AssetAccountService{repo: r, journalRepo: jr}
}

func (s *AssetAccountService) ListBalances(ctx context.Context) ([]AssetAccountBalance, error) {
	assets, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	bals, err := s.journalRepo.GetAccountBalancesAsOf(ctx, "", time.Now())
	if err != nil {
		return nil, err
	}
	balMap := map[int64]float64{}
	for _, b := range bals {
		balMap[b.AccountID] = b.Balance
	}
	res := make([]AssetAccountBalance, 0, len(assets))
	for _, a := range assets {
		res = append(res, AssetAccountBalance{
			AssetID:   a.ID,
			AccountID: a.AccountID,
			Balance:   balMap[a.AccountID],
		})
	}
	return res, nil
}

func (s *AssetAccountService) AdjustBalance(ctx context.Context, id int64, newBal float64) error {
	aa, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	bals, err := s.journalRepo.GetAccountBalancesAsOf(ctx, "", time.Now())
	if err != nil {
		return err
	}
	var cur float64
	for _, b := range bals {
		if b.AccountID == aa.AccountID {
			cur = b.Balance
			break
		}
	}
	diff := newBal - cur
	if diff == 0 {
		return nil
	}
	desc := fmt.Sprintf("Adjust asset %d balance", id)
	je := &models.JournalEntry{
		EntryDate:    time.Now(),
		Description:  &desc,
		SourceType:   "asset_adjust",
		SourceID:     fmt.Sprintf("%d", id),
		ShopUsername: "",
		Store:        "",
		CreatedAt:    time.Now(),
	}
	jid, err := s.journalRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	amt := math.Abs(diff)
	jl1 := &models.JournalLine{JournalID: jid, AccountID: aa.AccountID, IsDebit: diff > 0, Amount: amt}
	jl2 := &models.JournalLine{JournalID: jid, AccountID: 3001, IsDebit: diff < 0, Amount: amt}
	if err := s.journalRepo.InsertJournalLine(ctx, jl1); err != nil {
		return err
	}
	if err := s.journalRepo.InsertJournalLine(ctx, jl2); err != nil {
		return err
	}
	return nil
}
