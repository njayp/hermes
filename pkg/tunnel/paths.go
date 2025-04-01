package tunnel

import (
	"os"
)

var home = os.Getenv("HOME")
var certPath = home + "/cert.pem"
var ingressPath = home + "/ingress.yml"
var configPath = home + "/config.yml"

func withCert(args []string) []string {
	args = append(args, "--origincert", certPath)
	return args
}

func withConfig(args []string) []string {
	args = append(args, "--config", configPath)
	return args
}

func withLogLevel(args []string) []string {
	args = append(args, "--loglevel", "info") // change to debug for more logs
	return args
}
