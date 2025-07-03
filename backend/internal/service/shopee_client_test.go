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
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			refreshCalled = true
			r.ParseForm()
			if r.Form.Get("refresh_token") != "reftok" {
				t.Errorf("expected refresh_token=reftok, got %s", r.Form.Get("refresh_token"))
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
		PartnerID:     "pid",
		PartnerKey:    "secret",
		ShopID:        "shop",
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
		r.ParseForm()
		if r.URL.Query().Get("code") != "abc" {
			t.Errorf("query code=abc missing, got %s", r.URL.Query().Get("code"))
		}
		if r.Form.Get("code") != "abc" {
			t.Errorf("form code=abc missing, got %s", r.Form.Get("code"))
		}
		fmt.Fprint(w, `{"access_token":"tok"}`)
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     "pid",
		PartnerKey:    "secret",
	}
	c := NewShopeeClient(cfg)

	tok, err := c.GetAccessToken(context.Background(), "abc", "shop")
	if err != nil {
		t.Fatalf("GetAccessToken err: %v", err)
	}
	if tok != "tok" {
		t.Fatalf("unexpected token %s", tok)
	}
	if !called {
		t.Fatal("token endpoint not called")
	}
}
