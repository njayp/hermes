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

	cmd := exec.Command("cloudflared", "tunnel", "--origincert", certPath, "create", "--output", "json", name)
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
	args := []string{"tunnel"}
	args = append(args, "--origincert", certPath)
	args = append(args, "--config", configPath)
	args = append(args, "--grace-period", "5s")
	//args = append(args, "--loglevel", "debug")
	args = append(args, "run")
	args = append(args, t.Name)
	cmd := exec.CommandContext(ctx, "cloudflared", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd, cmd.Start()
}

func (t *Tunnel) Delete() {
	log.Printf("Deleting tunnel: %s", t.Name)
	exec.Command("cloudflared", "tunnel", "--origincert", certPath, "cleanup", t.Name).Run()
	exec.Command("cloudflared", "tunnel", "--origincert", certPath, "delete", t.Name).Run()
}
