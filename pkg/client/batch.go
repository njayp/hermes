package client

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

func (c *Client) BatchAddCNAME(ctx context.Context, reqs []*AddCNAMERequest) ([]*DelCNAMERequest, error) {
	ch := make(chan *DelCNAMERequest, len(reqs))
	wg := sync.WaitGroup{}
	for _, req := range reqs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Debug("creating dns record", "name", req.Name, "zone", req.ZoneID)
			// send create request
			record, err := c.AddCNAME(ctx, req)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			// save delete request
			ch <- &DelCNAMERequest{
				ZoneID:   req.ZoneID,
				RecordID: record.ID,
			}
		}()
	}

	// wait for all requests to finish
	wg.Wait()
	// close channel to prevent block
	close(ch)
	// collect all records
	records := make([]*DelCNAMERequest, 0, len(reqs))
	for record := range ch {
		records = append(records, record)
	}

	var err error
	// check if all records were created
	if len(records) != len(reqs) {
		err = fmt.Errorf("not all records were created, expected %d, got %d", len(reqs), len(records))
	}

	return records, err
}

// DelDNS deletes the DNS records created by AddDNS. It blocks until all
// records are deleted.
func (c *Client) BatchDelCNAME(ctx context.Context, records []*DelCNAMERequest) {
	wg := sync.WaitGroup{}
	for _, record := range records {
		wg.Add(1)
		// delete records async
		go func() {
			defer wg.Done()

			// send delete request
			_, err := c.DelCNAME(ctx, record)
			if err != nil {
				slog.Error(err.Error())
			}
		}()
	}
	wg.Wait()
}
