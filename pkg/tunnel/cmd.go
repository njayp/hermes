package tunnel

import (
	"context"
	"os"
	"os/exec"
)

var home = os.Getenv("HOME")
var certPath = home + "/cert.pem"
var ingressPath = home + "/ingress.yml"
var configPath = home + "/config.yml"

type tunCmd struct {
	tunnel         *Tunnel
	tunnelArgs     []string
	subCommand     string
	subCommandArgs []string
}

type tunCmdOpts = func(*tunCmd)

func newTunCmd(ctx context.Context, tunnel *Tunnel, subCmd string, opts ...tunCmdOpts) *exec.Cmd {
	c := &tunCmd{
		tunnel:     tunnel,
		subCommand: subCmd,
	}

	for _, opt := range opts {
		opt(c)
	}

	args := []string{"tunnel"}
	args = append(args, c.tunnelArgs...)
	args = append(args, c.subCommand)
	args = append(args, c.subCommandArgs...)
	args = append(args, c.tunnel.Name)
	return exec.CommandContext(ctx, "cloudflared", args...)
}

func withCert(c *tunCmd) {
	c.tunnelArgs = append(c.tunnelArgs, "--origincert", certPath)
}

func withConfig(c *tunCmd) {
	c.tunnelArgs = append(c.tunnelArgs, "--config", configPath)
}

func withLogLevel(c *tunCmd) {
	c.tunnelArgs = append(c.tunnelArgs, "--loglevel", "debug")
}

func withJson(c *tunCmd) {
	c.subCommandArgs = append(c.subCommandArgs, "--output", "json")
}
