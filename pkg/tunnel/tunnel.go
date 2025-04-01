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

func NewTunnel() (*Tunnel, error) {
	name := kvothe.GetRandomName()
	slog.Info(fmt.Sprintf("Creating tunnel: %s", name))

	args := []string{"tunnel"}
	args = withCert(args)
	args = append(args, "create", "--output", "json", name)
	cmd := exec.Command("cloudflared", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error creating tunnel: %w", err)
	}

	t := &Tunnel{}
	err = json.Unmarshal(out, t)
	if err != nil {
		slog.Error("error unmarshaling tunnel", slog.String("output", string(out)))
		return nil, fmt.Errorf("error unmarshaling tunnel: %w", err)
	}

	return t, nil
}

func (t *Tunnel) Run(ctx context.Context) (*exec.Cmd, error) {
	log.Printf("Running tunnel: %s", t.Name)

	// construct args
	args := []string{"tunnel"}
	args = withCert(args)
	args = withConfig(args)
	args = withLogLevel(args)
	args = append(args, "run", t.Name)
	cmd := exec.CommandContext(ctx, "cloudflared", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd, cmd.Start()
}

func (t *Tunnel) Delete() {
	slog.Info("Deleting tunnel", slog.String("name", t.Name))
	args := []string{"tunnel"}
	args = withCert(args)
	args = append(args, "cleanup", t.Name)
	err := exec.Command("cloudflared", args...).Run()
	if err != nil {
		slog.Error("error cleaning up tunnel", slog.String("name", t.Name), slog.String("error", err.Error()))
	}

	args = []string{"tunnel"}
	args = withCert(args)
	args = append(args, "delete", t.Name)
	err = exec.Command("cloudflared", args...).Run()
	if err != nil {
		slog.Error("error deleting tunnel", slog.String("name", t.Name), slog.String("error", err.Error()))
	}
}
