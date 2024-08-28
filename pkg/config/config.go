package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DiameterConfig DiameterConfig `json:"diameter"`
	RadiusConfig   RadiusConfig   `json:"radius"`
}

type DiameterConfig struct {
	Addr        string `json:"addr"`
	SSL         bool   `json:"ssl"`
	DiamHost    string `json:"diam_host"`
	DiamRealm   string `json:"diam_realm"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
	NetworkType string `json:"network_type"`
	// PeerAddr    string `json:"peer_addr"`
}

type RadiusConfig struct {
	Addr   string `json:"addr"`
	Secret string `json:"secret"`
	// ClientPort int    `json:"client_port"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
