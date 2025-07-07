package service

import (
	"context"
	"encoding/json"
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
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			refreshCalled = true
			if r.URL.Query().Get("partner_id") != "123" {
				t.Errorf("missing partner_id query")
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("invalid json: %v", err)
			}
			if payload["partner_id"] != float64(123) {
				t.Errorf("missing partner_id body")
			}
			if payload["refresh_token"] != "reftok" {
				t.Errorf("expected refresh_token=reftok, got %v", payload["refresh_token"])
			}
			if payload["shop_id"] != float64(456) {
				t.Errorf("expected shop_id=456, got %v", payload["shop_id"])
			}
			fmt.Fprint(w, `{"response":{"access_token":"newtoken"}}`)
		case "/api/v2/order/get_order_detail":
			detailCalled = true
			if r.URL.Query().Get("access_token") != "newtoken" {
				t.Errorf("expected access_token=newtoken, got %s", r.URL.Query().Get("access_token"))
			}
			if r.URL.Query().Get("order_sn_list") != "123" {
				t.Errorf("expected order_sn_list=123, got %s", r.URL.Query().Get("order_sn_list"))
			}
			fmt.Fprint(w, `{"response":{"order_status":"COMPLETE"}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     123,
		PartnerKey:    "deadbeef",
		ShopID:        456,
		AccessToken:   "oldtoken",
		RefreshToken:  "reftok",
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

func TestShopeeClientGetAccessTokenIncludesBody(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/auth/token/get" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		called = true
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("partner_id") != "123" {
			t.Errorf("missing partner_id query")
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("invalid json: %v", err)
		}
		if payload["code"] != "abc" {
			t.Errorf("code value missing")
		}
		if payload["shop_id"] != "456" {
			t.Errorf("shop_id value missing")
		}
		fmt.Fprint(w, `{"access_token":"tok"}`)
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     123,
		PartnerKey:    "deadbeef",
	}
	c := NewShopeeClient(cfg)

	resp, err := c.GetAccessToken(context.Background(), "abc", 456)
	if err != nil {
		t.Fatalf("GetAccessToken err: %v", err)
	}
	if resp.AccessToken != "tok" {
		t.Fatalf("unexpected token %s", resp.AccessToken)
	}
	if !called {
		t.Fatal("token endpoint not called")
	}
}
