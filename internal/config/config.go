package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
}

type ServerConfig struct {
	Address       string `env:"ADDRESS"`
	StoreInterval string `env:"STORE_INTERVAL"`
	FilePath      string `env:"STORE_FILE"`
	Restore       bool   `env:"RESTORE"`
}

func (c *AgentConfig) Parse() error {
	flag.StringVar(&c.Address, "a", "localhost:8080", "")
	flag.StringVar(&c.ReportInterval, "r", "10s", "")
	flag.StringVar(&c.PollInterval, "p", "2s", "")
	flag.Parse()
	err := env.Parse(c)
	return err
}

func (c *ServerConfig) Parse() error {
	flag.StringVar(&c.Address, "a", ":8080", "")
	flag.StringVar(&c.StoreInterval, "i", "300s", "")
	flag.StringVar(&c.FilePath, "f", "/tmp/devops-metrics-db.json", "")
	flag.BoolVar(&c.Restore, "r", true, "")
	flag.Parse()

	err := env.Parse(c)
	return err
}
