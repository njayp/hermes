package tunnel

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// ingressEntry represents each entry under the "ingress" key.
type ingressEntry struct {
	Hostname string `yaml:"hostname,omitempty"`
	Path     string `yaml:"path,omitempty"`
	Service  string `yaml:"service"`
}

// loadIngress reads the YAML file at the given path and returns a slice of IngressEntry. Panics on error.
func loadIngress() []ingressEntry {
	// Read the YAML file
	data, err := os.ReadFile(ingressPath)
	if err != nil {
		log.Fatalf("error reading YAML file: %v", err)
	}

	// Create an instance of the struct to hold the decoded data
	var ingress []ingressEntry

	// Unmarshal the YAML into the struct
	err = yaml.Unmarshal(data, &ingress)
	if err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	return ingress
}
