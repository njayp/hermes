package tunnel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/njayp/hermes/pkg/client"
	"gopkg.in/yaml.v2"
)

type TunnelConfig struct {
	Id        string         `yaml:"tunnel"`
	CredsPath string         `yaml:"credentials-file"`
	Ingress   []ingressEntry `yaml:"ingress"`

	// dns client
	cli *client.Client
}

func NewTunnelConfig(id string) TunnelConfig {
	return TunnelConfig{
		Id:        id,
		CredsPath: home + "/" + id + ".json",
		Ingress:   loadIngress(),
		cli:       client.NewClient(),
	}
}

func (c TunnelConfig) Save() error {
	// Marshal the struct into YAML.
	data, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	// Write the YAML data to a file named "config.yaml".
	return os.WriteFile(configPath, data, 0644)
}

func (c TunnelConfig) AddDNS(ctx context.Context) ([]*client.DelCNAMERequest, error) {
	zoneIdMap := make(map[string]string)
	reqs := []*client.AddCNAMERequest{}
	// process ingress entries
	for _, ie := range c.Ingress {
		split := strings.Split(ie.Hostname, ".")
		if len(split) < 3 {
			continue
		}
		name := split[0]
		host := split[1] + "." + split[2]

		// get zone id
		zoneId, ok := zoneIdMap[host]
		if !ok {
			zone, err := c.cli.GetZoneID(ctx, &client.GetZoneIDRequest{
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
			Content: c.Id + ".cfargotunnel.com",
			ZoneID:  zoneId,
		})
	}

	// create records async
	ch := make(chan *client.DelCNAMERequest, len(c.Ingress))
	for _, req := range reqs {
		go func() {
			slog.Debug("creating dns record", "name", req.Name, "zone", req.ZoneID)
			// send create request
			record, err := c.cli.AddCNAME(ctx, req)
			if err != nil {
				slog.Error(err.Error())
				ch <- nil
				return
			}

			// send delete request
			ch <- &client.DelCNAMERequest{
				ZoneID:   req.ZoneID,
				RecordID: record.ID,
			}
		}()
	}

	// records array doubles as wait group
	records := make([]*client.DelCNAMERequest, 0, len(reqs))
	ok := true
	for range reqs {
		record := <-ch
		if record == nil {
			ok = false
			continue
		}
		records = append(records, record)
	}
	if !ok {
		c.DelDNS(ctx, records)
		return nil, fmt.Errorf("failed to create dns records")
	}

	return records, nil
}

// DelDNS deletes the DNS records created by AddDNS. It blocks until all
// records are deleted.
func (c *TunnelConfig) DelDNS(ctx context.Context, records []*client.DelCNAMERequest) {
	wg := sync.WaitGroup{}
	for _, record := range records {
		wg.Add(1)
		// delete records async
		go func() {
			defer wg.Done()

			// send delete request
			_, err := c.cli.DelCNAME(ctx, record)
			if err != nil {
				slog.Error(err.Error())
			}
		}()
	}
	wg.Wait()
}
