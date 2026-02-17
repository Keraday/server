package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Cfg struct {
	ENV      string
	LogLvl   string
	ServAddr string
	DBURL    string
}

func Config() *Cfg {
	if os.Getenv("ENV") == "" {
		err := godotenv.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "WARNING: Using default env")
		}
	}
	return &Cfg{
		ENV:      getenv("ENV", "prod"),
		LogLvl:   getenv("LOG_LVL", "info"),
		ServAddr: getenv("SERV_ADDR", "localhost:8080"),
		DBURL:    mustGetenv("DB_URL"),
	}
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "error: required environment variable %s not set\n", key)
		os.Exit(1)
	}
	return val
}
