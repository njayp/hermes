package tunnel

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type TunnelConfig struct {
	Id        string         `yaml:"tunnel"`
	CredsPath string         `yaml:"credentials-file"`
	Ingress   []ingressEntry `yaml:"ingress"`
}

func NewTunnelConfig(id string) TunnelConfig {
	return TunnelConfig{
		Id:        id,
		CredsPath: home + "/" + id + ".json",
		Ingress:   loadIngress(),
	}
}

func (c TunnelConfig) WriteFile() error {
	// Marshal the struct into YAML.
	data, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	// Write the YAML data to a file named "config.yaml".
	return os.WriteFile(configPath, data, 0644)
}

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
