package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"time"
)

type AgentConfig struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	Key            string        `env:"KEY"`
}

type ServerConfig struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	FilePath      string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	DSN           string        `env:"DATABASE_DSN"`
}

func (c *AgentConfig) Parse() error {
	flag.StringVar(&c.Address, "a", "localhost:8080", "")
	flag.DurationVar(&c.ReportInterval, "r", 10*time.Second, "")
	flag.DurationVar(&c.PollInterval, "p", 2*time.Second, "")
	flag.StringVar(&c.Key, "k", "", "")
	flag.Parse()
	err := env.Parse(c)
	return err
}

func (c *ServerConfig) Parse() error {
	flag.StringVar(&c.Address, "a", ":8080", "")
	flag.DurationVar(&c.StoreInterval, "i", 300*time.Second, "")
	flag.StringVar(&c.FilePath, "f", "/tmp/devops-metrics-db.json", "")
	flag.BoolVar(&c.Restore, "r", true, "")
	flag.StringVar(&c.Key, "k", "", "")
	flag.StringVar(&c.DSN, "d", "", "")
	flag.Parse()

	err := env.Parse(c)
	return err
}
