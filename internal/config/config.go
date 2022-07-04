package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
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

func LoadServerConfig() (conf ServerConfig, err error) {
	flag.StringVar(&conf.Address, "a", ":8080", "")
	flag.DurationVar(&conf.StoreInterval, "i", 300*time.Second, "")
	flag.StringVar(&conf.FilePath, "f", "/tmp/devops-metrics-db.json", "")
	flag.BoolVar(&conf.Restore, "r", true, "")
	flag.StringVar(&conf.Key, "k", "", "")
	flag.StringVar(&conf.DSN, "d", "", "")
	flag.Parse()

	err = env.Parse(&conf)
	if err != nil {
		log.Fatal(err)
	}
	return
}
