package client

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/zones"
	"github.com/njayp/limiter"
)

type Client struct {
	api     *cloudflare.Client
	limiter *limiter.Limiter
}

func NewClient() *Client {
	return &Client{
		api: cloudflare.NewClient(),
		// cloudflare api rate limits are 1200 requests per 5 minutes
		// limiter also staggers requests by 50ms
		limiter: limiter.NewLimiter(1200, time.Minute*5, time.Millisecond*50),
	}
}

func (c *Client) AddCNAME(ctx context.Context, req AddCNAMERequest) (*dns.RecordResponse, error) {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return c.api.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.String(req.ZoneID),
		Record: dns.CNAMERecordParam{
			Name:    cloudflare.String(req.Name),
			Content: cloudflare.String(req.Content),
			Proxied: cloudflare.Bool(true),
			TTL:     cloudflare.Raw[dns.TTL](1),
			Type:    cloudflare.Raw[dns.CNAMERecordType](dns.CNAMERecordTypeCNAME),
		},
	})
}

func (c *Client) DelCNAME(ctx context.Context, req DelCNAMERequest) (*dns.RecordDeleteResponse, error) {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return c.api.DNS.Records.Delete(ctx, req.RecordID, dns.RecordDeleteParams{
		ZoneID: cloudflare.String(req.ZoneID),
	})
}

func (c *Client) GetZoneID(ctx context.Context, req GetZoneIDRequest) (*zones.Zone, error) {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	list, err := c.api.Zones.List(ctx, zones.ZoneListParams{
		Name: cloudflare.String(req.Name),
	})
	if err != nil {
		return nil, err
	}
	if len(list.Result) == 0 {
		return nil, fmt.Errorf("zone not found")
	}

	// assuming only one zone is returned
	return &list.Result[0], err
}
