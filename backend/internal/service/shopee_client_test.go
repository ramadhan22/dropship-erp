package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestShopeeClientRefreshAndGetOrderDetail(t *testing.T) {
	var refreshCalled, detailCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/access_token/get":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			refreshCalled = true
			if r.URL.Query().Get("partner_id") != "12345" {
				t.Errorf("missing partner_id query")
			}
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("invalid json: %v", err)
			}
			if payload["refresh_token"] != "reftok" {
				t.Errorf("expected refresh_token=reftok, got %v", payload["refresh_token"])
			}
			if payload["shop_id"] != float64(2) {
				t.Errorf("expected shop_id=2, got %v", payload["shop_id"])
			}
			fmt.Fprint(w, `{"access_token":"newtoken"}`)
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

	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(srv.URL, "http://")
		return oldTransport.RoundTrip(req)
	})
	defer func() { http.DefaultTransport = oldTransport }()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     "12345",
		PartnerKey:    "secret",
		ShopID:        "2",
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
		if r.URL.Query().Get("partner_id") != "12345" {
			t.Errorf("missing partner_id query")
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("invalid json: %v", err)
		}
		if payload["code"] != "abc" {
			t.Errorf("code value missing")
		}
		if payload["shop_id"] != "2" {
			t.Errorf("shop_id value missing")
		}
		fmt.Fprint(w, `{"access_token":"tok"}`)
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     "12345",
		PartnerKey:    "secret",
	}
	c := NewShopeeClient(cfg)

	resp, err := c.GetAccessToken(context.Background(), "abc", "2")
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
