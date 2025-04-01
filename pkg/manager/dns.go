package manager

import (
	"context"
	"strings"

	"github.com/njayp/hermes/pkg/client"
	"github.com/njayp/hermes/pkg/tunnel"
)

var cli = client.NewClient()

func makeDNSRequests(ctx context.Context, conf tunnel.TunnelConfig) ([]*client.AddCNAMERequest, error) {
	zoneIdMap := make(map[string]string)
	reqs := []*client.AddCNAMERequest{}
	// process ingress entries
	for _, ie := range conf.Ingress {
		split := strings.Split(ie.Hostname, ".")
		// skip if not a valid hostname
		if len(split) < 3 {
			continue
		}
		name := split[0]
		host := split[1] + "." + split[2]

		// get zone id
		zoneId, ok := zoneIdMap[host]
		if !ok {
			zone, err := cli.GetZoneID(ctx, &client.GetZoneIDRequest{
				Name: host,
			})
			if err != nil {
				return nil, err
			}

			zoneId = zone.ID
			zoneIdMap[host] = zone.ID
		}

		// create record request
		reqs = append(reqs, &client.AddCNAMERequest{
			Name:    name,
			Content: conf.Id + ".cfargotunnel.com",
			ZoneID:  zoneId,
		})
	}

	return reqs, nil
}

func addDNS(ctx context.Context, conf tunnel.TunnelConfig) ([]*client.DelCNAMERequest, error) {
	// make dns requests from ingress entries
	reqs, err := makeDNSRequests(ctx, conf)
	if err != nil {
		return nil, err
	}

	records, err := cli.BatchAddCNAME(ctx, reqs)
	if err != nil {
		// cleanup created dns records
		cli.BatchDelCNAME(ctx, records)
		return nil, err
	}

	return records, nil
}
