package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Database struct {
	URI string
}

type Measurement struct {
	ListenAddress string
	ImportDir     string
	Interval      time.Duration
	Endpoint      string
}

type Dashboard struct {
	ListenAddress string
	Limit         int
}

func LoadDatabase() (Database, error) {
	uri := os.Getenv("NETZBREMSE_DATABASE_URL")
	if uri == "" {
		uri = os.Getenv("DATABASE_URL")
	}
	if uri == "" {
		return Database{}, fmt.Errorf("NETZBREMSE_DATABASE_URL or DATABASE_URL is required")
	}
	return Database{URI: uri}, nil
}

func LoadMeasurement() (Measurement, error) {
	interval := 1 * time.Hour
	if raw := os.Getenv("NETZBREMSE_MEASUREMENT_INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return Measurement{}, fmt.Errorf("parse NETZBREMSE_MEASUREMENT_INTERVAL: %w", err)
		}
		interval = parsed
	}
	return Measurement{
		ListenAddress: envOrDefault("NETZBREMSE_MEASUREMENT_LISTEN_ADDR", ":8081"),
		ImportDir:     os.Getenv("NETZBREMSE_IMPORT_DIR"),
		Interval:      interval,
		Endpoint:      envOrDefault("NETZBREMSE_ENDPOINT", "https://speedtest.m0sh1.cc"),
	}, nil
}

func LoadDashboard() (Dashboard, error) {
	limit := 50
	if raw := os.Getenv("NETZBREMSE_DASHBOARD_LIMIT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Dashboard{}, fmt.Errorf("parse NETZBREMSE_DASHBOARD_LIMIT: %w", err)
		}
		limit = parsed
	}
	return Dashboard{
		ListenAddress: envOrDefault("NETZBREMSE_DASHBOARD_LISTEN_ADDR", ":8501"),
		Limit:         limit,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
