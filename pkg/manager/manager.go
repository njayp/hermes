package manager

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/njayp/hermes/pkg/tunnel"
)

func Run(ctx context.Context) error {
	// create tunnel
	tun, err := tunnel.NewTunnel()
	if err != nil {
		return err
	}
	// defer is first in, last out
	defer tun.Delete()

	// save config
	conf := tunnel.NewTunnelConfig(tun.Id)
	err = conf.Save()
	if err != nil {
		return err
	}
	// add dns entries
	records, err := conf.AddDNS(ctx)
	if err != nil {
		return err
	}
	// rm dns entries
	defer conf.DelDNS(ctx, records)

	// run tunnel
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	cmd, err := tun.Run(ctx)
	if err != nil {
		return err
	}
	// wait for the tunnel to close before cleaning up
	defer cmd.Wait()
	// cleanup if the tunnel exits early
	go func() {
		cmd.Wait()
		cancel(fmt.Errorf("tunnel exited"))
	}()

	// Wait for signal
	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case sig := <-intCh:
		err := fmt.Errorf("received signal: %s", sig)
		cancel(err)
		return err
	}
}
