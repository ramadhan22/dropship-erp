package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/ramadhan22/dropship-erp/backend/internal/handlers"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// Mock channel service that implements all required methods
type mockChannelService struct{}

func (m *mockChannelService) CreateJenisChannel(ctx context.Context, jenisChannel string) (int64, error) {
	return 1, nil
}

func (m *mockChannelService) CreateStore(ctx context.Context, channelID int64, namaToko string) (int64, error) {
	return 1, nil
}

func (m *mockChannelService) ListJenisChannels(ctx context.Context) ([]models.JenisChannel, error) {
	return []models.JenisChannel{}, nil
}

func (m *mockChannelService) ListStoresByChannel(ctx context.Context, channelID int64) ([]models.Store, error) {
	return []models.Store{}, nil
}

func (m *mockChannelService) ListStoresByChannelName(ctx context.Context, channelName string) ([]models.Store, error) {
	return []models.Store{}, nil
}

func (m *mockChannelService) GetStore(ctx context.Context, id int64) (*models.Store, error) {
	return &models.Store{}, nil
}

func (m *mockChannelService) ListAllStores(ctx context.Context) ([]models.StoreWithChannel, error) {
	// Return mock data that matches what should be in the database
	return []models.StoreWithChannel{
		{
			Store: models.Store{
				StoreID:        1,
				JenisChannelID: 1,
				NamaToko:       "MR eStore Shopee",
			},
			JenisChannel: "Shopee",
		},
		{
			Store: models.Store{
				StoreID:        2,
				JenisChannelID: 1,
				NamaToko:       "MR Barista Gear",
			},
			JenisChannel: "Shopee",
		},
	}, nil
}

func (m *mockChannelService) UpdateStore(ctx context.Context, st *models.Store) error {
	return nil
}

func (m *mockChannelService) DeleteStore(ctx context.Context, id int64) error {
	return nil
}

func main() {
	fmt.Println("=== Testing /api/stores/all endpoint ===")
	
	gin.SetMode(gin.TestMode)
	
	// Create mock service
	mockSvc := &mockChannelService{}
	
	// Create handler
	handler := handlers.NewChannelHandler(mockSvc)
	
	// Create router and route
	router := gin.New()
	router.GET("/api/stores/all", handler.HandleListAllStores)
	
	// Create request
	req := httptest.NewRequest("GET", "/api/stores/all", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	router.ServeHTTP(w, req)
	
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())
	
	if w.Code != http.StatusOK {
		fmt.Printf("ERROR: Expected status 200, got %d\n", w.Code)
		return
	}
	
	// Parse response as frontend would
	var stores []struct {
		StoreID  int64  `json:"store_id"`
		NamaToko string `json:"nama_toko"`
	}
	
	err := json.Unmarshal(w.Body.Bytes(), &stores)
	if err != nil {
		fmt.Printf("ERROR: Failed to parse JSON: %v\n", err)
		return
	}
	
	fmt.Printf("SUCCESS: Frontend would see %d stores:\n", len(stores))
	for i, store := range stores {
		fmt.Printf("  Store %d: ID=%d, Name=%s\n", i+1, store.StoreID, store.NamaToko)
	}
	
	// Test what the frontend dropdown would show
	fmt.Println("\n=== Frontend Dropdown Options ===")
	for _, store := range stores {
		fmt.Printf("  <MenuItem value=%d>%s</MenuItem>\n", store.StoreID, store.NamaToko)
	}
}