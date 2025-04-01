package tunnel

import (
	"os"
)

var home = os.Getenv("HOME")
var certPath = home + "/cert.pem"
var ingressPath = home + "/ingress.yml"
var configPath = home + "/config.yml"
var tunnelPath = home + "/tunnel.yml"
