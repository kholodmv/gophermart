package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	Env                  string
}

func UseServerStartParams() Config {
	var c Config

	flag.StringVar(&c.RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&c.DatabaseURI, "d", "", "connection string to postgres db")
	flag.StringVar(&c.AccrualSystemAddress, "r", "", "billing system address")
	flag.StringVar(&c.Env, "e", "dev", "environment")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		c.RunAddress = envRunAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		c.DatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		c.AccrualSystemAddress = envAccrualSystemAddress
	}
	if envEnv := os.Getenv("ENVIRONMENT"); envEnv != "" {
		c.Env = envEnv
	}

	return c
}
