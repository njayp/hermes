package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os/exec"

	"github.com/njayp/kvothe"
)

type Tunnel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewTunnel(ctx context.Context) (*Tunnel, error) {
	tun := &Tunnel{
		Name: kvothe.GetRandomName(),
	}

	slog.Info("Creating tunnel", slog.String("name", tun.Name))
	cmd := tun.newCmd(ctx, "create", withCert, withJson)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error creating tunnel: %w", err)
	}

	err = json.Unmarshal(out, tun)
	if err != nil {
		slog.Error("error unmarshaling tunnel", slog.String("output", string(out)))
		return nil, fmt.Errorf("error unmarshaling tunnel: %w", err)
	}

	return tun, nil
}

func (t *Tunnel) Run(ctx context.Context) (*exec.Cmd, error) {
	log.Printf("Running tunnel: %s", t.Name)
	cmd := t.newCmd(ctx, "run", withCert, withConfig, withLogLevel)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd, cmd.Start()
}

func (t *Tunnel) Delete() {
	slog.Info("Deleting tunnel", slog.String("name", t.Name))
	ctx := context.TODO()

	err := t.newCmd(ctx, "cleanup", withCert).Run()
	if err != nil {
		slog.Error("error cleaning up tunnel", slog.String("name", t.Name), slog.String("error", err.Error()))
	}

	err = t.newCmd(ctx, "delete", withCert).Run()
	if err != nil {
		slog.Error("error deleting tunnel", slog.String("name", t.Name), slog.String("error", err.Error()))
	}
}

func (t *Tunnel) newCmd(ctx context.Context, subCmd string, opts ...tunCmdOpts) *exec.Cmd {
	return newTunCmd(ctx, t, subCmd, opts...)
}
