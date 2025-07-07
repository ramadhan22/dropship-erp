package service

import (
	"context"
	"strings"
	"time"

	"github.com/teacat/shopeego"
)

// RefreshAccessTokenShopeeGo refreshes the access token using the shopeego library.
// This helper only works with the default Shopee endpoints since shopeego does
// not support custom base URLs.
func (c *ShopeeClient) RefreshAccessTokenShopeeGo(ctx context.Context) (*refreshResp, error) {
	opts := &shopeego.ClientOptions{
		Secret:    c.PartnerKey,
		IsSandbox: strings.Contains(c.BaseURL, "uat") || strings.Contains(c.BaseURL, "test"),
		Version:   shopeego.ClientVersionV2,
	}
	cli := shopeego.NewClient(opts)

	req := &shopeego.RefreshAccessTokenRequest{
		RefreshToken: c.RefreshToken,
		ShopID:       c.ShopID,
		PartnerID:    c.PartnerID,
		Timestamp:    int(time.Now().Unix()),
	}
	resp, err := cli.RefreshAccessToken(req)
	if err != nil {
		return nil, err
	}
	if resp.AccessToken != "" {
		c.AccessToken = resp.AccessToken
	}
	if resp.RefreshToken != "" {
		c.RefreshToken = resp.RefreshToken
	}

	out := &refreshResp{}
	out.Response.AccessToken = resp.AccessToken
	out.Response.RefreshToken = resp.RefreshToken
	out.Response.ExpireIn = resp.ExpireIn
	out.Response.RequestID = resp.RequestID
	out.Error = resp.Error
	return out, nil
}
