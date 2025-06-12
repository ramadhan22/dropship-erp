package service

import (
	"context"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ChannelRepoInterface defines repo methods used by the service.
type ChannelRepoInterface interface {
        CreateJenisChannel(ctx context.Context, c *models.JenisChannel) (int64, error)
        CreateStore(ctx context.Context, s *models.Store) (int64, error)
        ListJenisChannels(ctx context.Context) ([]models.JenisChannel, error)
        ListStoresByChannel(ctx context.Context, channelID int64) ([]models.Store, error)
        ListStoresByChannelName(ctx context.Context, channelName string) ([]models.Store, error)
}

// ChannelService provides master data operations for jenis_channels and stores.
type ChannelService struct {
	repo ChannelRepoInterface
}

// NewChannelService constructs a ChannelService.
func NewChannelService(r ChannelRepoInterface) *ChannelService {
	return &ChannelService{repo: r}
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
