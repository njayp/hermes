package tunnel

import (
	"context"
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

func (c TunnelConfig) AddDNS(ctx context.Context) ([]client.DelCNAMERequest, error) {
	zoneIdMap := make(map[string]string)
	recordsMap := make(map[string]client.AddCNAMERequest)
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
			zone, err := c.cli.GetZoneID(ctx, client.GetZoneIDRequest{
				Name: host,
			})
			if err != nil {
				// TODO gracefully exit
				return nil, err
			}

			zoneId = zone.ID
			zoneIdMap[host] = zone.ID
		}

		// create record request
		record := client.AddCNAMERequest{
			Name:    name,
			Content: c.Id + ".cfargotunnel.com",
			ZoneID:  zoneId,
		}
		recordsMap[host] = record
	}

	// create records async
	ch := make(chan client.DelCNAMERequest, len(c.Ingress))
	for _, v := range recordsMap {
		go func() {
			slog.Debug("creating dns record", "name", v.Name, "zone", v.ZoneID)
			// send create request
			record, err := c.cli.AddCNAME(ctx, v)
			if err != nil {
				// TODO gracefully exit
				slog.Error(err.Error())
			}

			// create delete request
			delRecord := client.DelCNAMERequest{
				ZoneID:   v.ZoneID,
				RecordID: record.ID,
			}

			// send to channel
			ch <- delRecord
		}()
	}

	// records array doubles as wait group
	records := make([]client.DelCNAMERequest, 0, len(recordsMap))
	for range recordsMap {
		records = append(records, <-ch)
	}

	return records, nil
}

// DelDNS deletes the DNS records created by AddDNS. It blocks until all
// records are deleted.
func (c *TunnelConfig) DelDNS(ctx context.Context, records []client.DelCNAMERequest) {
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
