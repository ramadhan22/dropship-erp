package service

import (
	"encoding/json"
	"testing"
)

// TestShopeeAdsPerformanceResponse_JSONUnmarshaling verifies that the struct
// can correctly unmarshal the Shopee API response format (single object, not array)
func TestShopeeAdsPerformanceResponse_JSONUnmarshaling(t *testing.T) {
	// This test ensures the fix for issue #390 works correctly
	// The API returns response as a single object, not an array
	jsonData := `{
		"response": {
			"shop_id": 38526489,
			"region": "SG",
			"campaign_list": [
				{
					"campaign_id": 298705350,
					"ad_type": "product",
					"campaign_placement": "search",
					"ad_name": "Test Campaign",
					"metrics_list": [
						{
							"hour": 0,
							"date": "17-07-2025",
							"impression": 100,
							"clicks": 5,
							"ctr": 0.05,
							"expense": 500.0,
							"broad_gmv": 1000.0,
							"broad_order": 2,
							"broad_order_amount": 1000.0,
							"broad_roi": 2.0,
							"broad_cir": 0.5,
							"cr": 0.4,
							"cpc": 100.0,
							"direct_order": 2,
							"direct_order_amount": 1000.0,
							"direct_gmv": 1000.0,
							"direct_roi": 2.0,
							"direct_cir": 0.5,
							"direct_cr": 0.4,
							"cpdc": 250.0
						}
					]
				}
			]
		},
		"error": "",
		"message": "",
		"warning": "",
		"request_id": "test-request-123"
	}`

	var resp ShopeeAdsPerformanceResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify the basic structure
	if resp.Response.ShopID != 38526489 {
		t.Errorf("Expected shop_id 38526489, got %d", resp.Response.ShopID)
	}

	if resp.Response.Region != "SG" {
		t.Errorf("Expected region 'SG', got '%s'", resp.Response.Region)
	}

	if len(resp.Response.CampaignList) != 1 {
		t.Errorf("Expected 1 campaign, got %d", len(resp.Response.CampaignList))
	}

	// Verify campaign details
	campaign := resp.Response.CampaignList[0]
	if campaign.CampaignID != 298705350 {
		t.Errorf("Expected campaign_id 298705350, got %d", campaign.CampaignID)
	}

	if len(campaign.MetricsList) != 1 {
		t.Errorf("Expected 1 metrics entry, got %d", len(campaign.MetricsList))
	}

	// Verify metrics details
	metrics := campaign.MetricsList[0]
	if metrics.Hour != 0 {
		t.Errorf("Expected hour 0, got %d", metrics.Hour)
	}

	if metrics.Date != "17-07-2025" {
		t.Errorf("Expected date '17-07-2025', got '%s'", metrics.Date)
	}

	if metrics.Impression != 100 {
		t.Errorf("Expected impression 100, got %d", metrics.Impression)
	}

	if metrics.Clicks != 5 {
		t.Errorf("Expected clicks 5, got %d", metrics.Clicks)
	}

	if resp.RequestID != "test-request-123" {
		t.Errorf("Expected request_id 'test-request-123', got '%s'", resp.RequestID)
	}
}

// TestShopeeAdsPerformanceResponse_EmptyResponse verifies the struct can handle empty campaign lists
func TestShopeeAdsPerformanceResponse_EmptyResponse(t *testing.T) {
	jsonData := `{
		"response": {
			"shop_id": 12345,
			"region": "MY",
			"campaign_list": []
		},
		"error": "",
		"message": "No campaigns found",
		"warning": "",
		"request_id": "empty-test-456"
	}`

	var resp ShopeeAdsPerformanceResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty response JSON: %v", err)
	}

	if resp.Response.ShopID != 12345 {
		t.Errorf("Expected shop_id 12345, got %d", resp.Response.ShopID)
	}

	if len(resp.Response.CampaignList) != 0 {
		t.Errorf("Expected empty campaign list, got %d campaigns", len(resp.Response.CampaignList))
	}

	if resp.Message != "No campaigns found" {
		t.Errorf("Expected message 'No campaigns found', got '%s'", resp.Message)
	}
}

// TestShopeeAdsPerformanceResponse_ErrorResponse verifies the struct can handle API error responses
func TestShopeeAdsPerformanceResponse_ErrorResponse(t *testing.T) {
	jsonData := `{
		"response": {
			"shop_id": 0,
			"region": "",
			"campaign_list": []
		},
		"error": "INVALID_ACCESS_TOKEN",
		"message": "Access token is invalid or expired",
		"warning": "",
		"request_id": "error-test-789"
	}`

	var resp ShopeeAdsPerformanceResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response JSON: %v", err)
	}

	if resp.Error != "INVALID_ACCESS_TOKEN" {
		t.Errorf("Expected error 'INVALID_ACCESS_TOKEN', got '%s'", resp.Error)
	}

	if resp.Message != "Access token is invalid or expired" {
		t.Errorf("Expected message 'Access token is invalid or expired', got '%s'", resp.Message)
	}

	if resp.RequestID != "error-test-789" {
		t.Errorf("Expected request_id 'error-test-789', got '%s'", resp.RequestID)
	}
}