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
	// cleanup if the tunnel exits early
	go func() {
		cmd.Wait()
		cancel(fmt.Errorf("tunnel exited"))
	}()
	// we should wait for the tunnel process to exit
	// but cloudflared often hangs on interrupt
	//defer cmd.Wait()

	// Wait for signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case sig := <-ch:
		err := fmt.Errorf("received signal: %s", sig)
		cancel(err)
		return err
	}
}
