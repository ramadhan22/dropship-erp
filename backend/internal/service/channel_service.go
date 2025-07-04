package service

import (
	"context"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ChannelRepoInterface defines repo methods used by the service.
type ChannelRepoInterface interface {
	CreateJenisChannel(ctx context.Context, c *models.JenisChannel) (int64, error)
	CreateStore(ctx context.Context, s *models.Store) (int64, error)
	ListJenisChannels(ctx context.Context) ([]models.JenisChannel, error)
	ListStoresByChannel(ctx context.Context, channelID int64) ([]models.Store, error)
	ListStoresByChannelName(ctx context.Context, channelName string) ([]models.Store, error)
	GetStoreByID(ctx context.Context, id int64) (*models.Store, error)
	GetStoreByName(ctx context.Context, name string) (*models.Store, error)
	ListAllStores(ctx context.Context) ([]models.StoreWithChannel, error)
	UpdateStore(ctx context.Context, s *models.Store) error
	DeleteStore(ctx context.Context, id int64) error
}

// ChannelService provides master data operations for jenis_channels and stores.
type ChannelService struct {
	repo   ChannelRepoInterface
	client *ShopeeClient
}

// NewChannelService constructs a ChannelService.
func NewChannelService(r ChannelRepoInterface, c *ShopeeClient) *ChannelService {
	return &ChannelService{repo: r, client: c}
}

func (s *ChannelService) CreateJenisChannel(ctx context.Context, jenis string) (int64, error) {
	c := &models.JenisChannel{JenisChannel: jenis}
	return s.repo.CreateJenisChannel(ctx, c)
}

func (s *ChannelService) CreateStore(ctx context.Context, channelID int64, namaToko string) (int64, error) {
	st := &models.Store{JenisChannelID: channelID, NamaToko: namaToko}
	return s.repo.CreateStore(ctx, st)
}

func (s *ChannelService) ListJenisChannels(ctx context.Context) ([]models.JenisChannel, error) {
	return s.repo.ListJenisChannels(ctx)
}

func (s *ChannelService) ListStoresByChannel(ctx context.Context, channelID int64) ([]models.Store, error) {
	return s.repo.ListStoresByChannel(ctx, channelID)
}

func (s *ChannelService) ListStoresByChannelName(ctx context.Context, channelName string) ([]models.Store, error) {
	return s.repo.ListStoresByChannelName(ctx, channelName)
}

func (s *ChannelService) GetStore(ctx context.Context, id int64) (*models.Store, error) {
	return s.repo.GetStoreByID(ctx, id)
}

func (s *ChannelService) GetStoreByName(ctx context.Context, name string) (*models.Store, error) {
	return s.repo.GetStoreByName(ctx, name)
}

func (s *ChannelService) ListAllStores(ctx context.Context) ([]models.StoreWithChannel, error) {
	return s.repo.ListAllStores(ctx)
}

func (s *ChannelService) UpdateStore(ctx context.Context, st *models.Store) error {
	if st.CodeID != nil && st.ShopID != nil && s.client != nil {
		tok, err := s.client.GetAccessToken(ctx, *st.CodeID, *st.ShopID)
		if err != nil {
			return err
		}
		st.AccessToken = &tok.AccessToken
		st.RefreshToken = &tok.RefreshToken
		st.ExpireIn = &tok.ExpireIn
		st.RequestID = &tok.RequestID
		now := time.Now()
		st.LastUpdated = &now
	}
	return s.repo.UpdateStore(ctx, st)
}

func (s *ChannelService) DeleteStore(ctx context.Context, id int64) error {
	return s.repo.DeleteStore(ctx, id)
}
