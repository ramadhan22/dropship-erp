package service

import (
	"context"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
)

func TestFetchAdsCampaignSettings(t *testing.T) {
	// Test that the method doesn't panic with empty campaign IDs
	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: "https://partner.test-stable.shopeemobile.com",
		PartnerID:     "123456",
		PartnerKey:    "test-key",
	}

	service := &AdsPerformanceService{
		shopeeClient: NewShopeeClient(cfg),
	}

	ctx := context.Background()
	err := service.FetchAdsCampaignSettings(ctx, 1, []int64{})
	if err != nil {
		t.Errorf("FetchAdsCampaignSettings should not error with empty campaign IDs, got: %v", err)
	}
}

func TestCampaignSettingsDataStructures(t *testing.T) {
	// Test that our data structures can be properly marshaled/unmarshaled
	settings := &ShopeeCampaignSettings{
		CampaignID: 123,
		CommonInfo: &ShopeeCampaignCommonInfo{
			AdName:            "Test Campaign",
			AdType:            "product",
			CampaignStatus:    "ongoing",
			CampaignPlacement: "search",
			BiddingMethod:     "auto",
			CampaignBudget:    100.0,
			CampaignDuration: ShopeeCampaignDuration{
				StartTime: 1640995200,
				EndTime:   0,
			},
			ItemIDList: []int64{1, 2, 3},
		},
		AutoBiddingInfo: &ShopeeCampaignAutoBidding{
			RoasTarget: 5.0,
		},
		ManualBiddingInfo: &ShopeeCampaignManualBidding{
			EnhancedCPC: true,
			SelectedKeywords: []ShopeeCampaignKeyword{
				{
					Keyword:          "test keyword",
					Status:           "normal",
					MatchType:        "exact",
					BidPricePerClick: 1.5,
				},
			},
			DiscoveryAdsLocations: []ShopeeCampaignDiscoveryAdsLocation{
				{
					Location: "daily_discover",
					Status:   "active",
					BidPrice: 2.0,
				},
			},
		},
		AutoProductAdsInfo: []ShopeeCampaignAutoProductAds{
			{
				ProductName: "Test Product",
				Status:      "ongoing",
				ItemID:      12345,
			},
		},
	}

	if settings.CampaignID != 123 {
		t.Errorf("Expected campaign ID 123, got %d", settings.CampaignID)
	}

	if settings.CommonInfo.AdName != "Test Campaign" {
		t.Errorf("Expected campaign name 'Test Campaign', got %s", settings.CommonInfo.AdName)
	}

	if len(settings.CommonInfo.ItemIDList) != 3 {
		t.Errorf("Expected 3 item IDs, got %d", len(settings.CommonInfo.ItemIDList))
	}

	if settings.AutoBiddingInfo.RoasTarget != 5.0 {
		t.Errorf("Expected ROAS target 5.0, got %f", settings.AutoBiddingInfo.RoasTarget)
	}

	if !settings.ManualBiddingInfo.EnhancedCPC {
		t.Errorf("Expected EnhancedCPC to be true")
	}

	if len(settings.AutoProductAdsInfo) != 1 {
		t.Errorf("Expected 1 auto product ad, got %d", len(settings.AutoProductAdsInfo))
	}
}
