package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
)

func TestShopeeClientRefreshAndGetOrderDetail(t *testing.T) {
	var refreshCalled, detailCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/access_token/get":
			refreshCalled = true
			fmt.Fprint(w, `{"response":{"access_token":"newtoken"}}`)
		case "/api/v2/order/get_order_detail":
			detailCalled = true
			if r.URL.Query().Get("access_token") != "newtoken" {
				t.Errorf("expected access_token=newtoken, got %s", r.URL.Query().Get("access_token"))
			}
			fmt.Fprint(w, `{"response":{"order_status":"COMPLETE"}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURL:      srv.URL,
		PartnerID:    "pid",
		PartnerKey:   "secret",
		ShopID:       "shop",
		AccessToken:  "oldtoken",
		RefreshToken: "reftok",
	}
	c := NewShopeeClient(cfg)

	status, err := c.GetOrderDetail(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetOrderDetail err: %v", err)
	}
	if status != "COMPLETE" {
		t.Fatalf("unexpected status %s", status)
	}
	if !refreshCalled {
		t.Fatal("RefreshAccessToken not called")
	}
	if !detailCalled {
		t.Fatal("order detail not fetched")
	}
	if c.AccessToken != "newtoken" {
		t.Fatalf("token not updated: %s", c.AccessToken)
	}
}
