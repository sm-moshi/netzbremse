package config

import (
	"testing"
	"time"
)

func TestLoadDatabase(t *testing.T) {
	t.Run("NETZBREMSE_DATABASE_URL takes precedence", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DATABASE_URL", "postgres://primary:5432/db")
		t.Setenv("DATABASE_URL", "postgres://fallback:5432/db")

		cfg, err := LoadDatabase()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.URI != "postgres://primary:5432/db" {
			t.Errorf("URI = %q, want %q", cfg.URI, "postgres://primary:5432/db")
		}
	})

	t.Run("falls back to DATABASE_URL", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DATABASE_URL", "")
		t.Setenv("DATABASE_URL", "postgres://fallback:5432/db")

		cfg, err := LoadDatabase()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.URI != "postgres://fallback:5432/db" {
			t.Errorf("URI = %q, want %q", cfg.URI, "postgres://fallback:5432/db")
		}
	})

	t.Run("error when both unset", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DATABASE_URL", "")
		t.Setenv("DATABASE_URL", "")

		_, err := LoadDatabase()
		if err == nil {
			t.Fatal("expected error when no database URL is set")
		}
	})
}

func TestLoadMeasurement(t *testing.T) {
	// Set required env so LoadDatabase won't interfere — LoadMeasurement doesn't
	// need it, but ensure a clean slate.

	t.Run("defaults", func(t *testing.T) {
		// Clear all measurement env vars.
		t.Setenv("NETZBREMSE_MEASUREMENT_INTERVAL", "")
		t.Setenv("NETZBREMSE_MEASUREMENT_LISTEN_ADDR", "")
		t.Setenv("NETZBREMSE_IMPORT_DIR", "")
		t.Setenv("NETZBREMSE_ENDPOINT", "")
		t.Setenv("NETZBREMSE_SPEEDTEST_COMMAND", "")
		t.Setenv("NETZBREMSE_SPEEDTEST_TIMEOUT", "")

		cfg, err := LoadMeasurement()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ListenAddress != ":8081" {
			t.Errorf("ListenAddress = %q, want %q", cfg.ListenAddress, ":8081")
		}
		if cfg.ImportDir != "" {
			t.Errorf("ImportDir = %q, want empty", cfg.ImportDir)
		}
		if cfg.Interval != 1*time.Hour {
			t.Errorf("Interval = %v, want %v", cfg.Interval, 1*time.Hour)
		}
		if cfg.Endpoint != "https://netzbremse.de/speed" {
			t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "https://netzbremse.de/speed")
		}
		if cfg.Command != "node /app/netzbremse-browser.mjs" {
			t.Errorf("Command = %q, want %q", cfg.Command, "node /app/netzbremse-browser.mjs")
		}
		if cfg.Timeout != 2*time.Minute {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 2*time.Minute)
		}
	})

	t.Run("custom interval", func(t *testing.T) {
		t.Setenv("NETZBREMSE_MEASUREMENT_INTERVAL", "30m")
		t.Setenv("NETZBREMSE_MEASUREMENT_LISTEN_ADDR", "")
		t.Setenv("NETZBREMSE_IMPORT_DIR", "")
		t.Setenv("NETZBREMSE_ENDPOINT", "")
		t.Setenv("NETZBREMSE_SPEEDTEST_COMMAND", "")
		t.Setenv("NETZBREMSE_SPEEDTEST_TIMEOUT", "")

		cfg, err := LoadMeasurement()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Interval != 30*time.Minute {
			t.Errorf("Interval = %v, want %v", cfg.Interval, 30*time.Minute)
		}
	})

	t.Run("invalid interval", func(t *testing.T) {
		t.Setenv("NETZBREMSE_MEASUREMENT_INTERVAL", "not-a-duration")
		t.Setenv("NETZBREMSE_MEASUREMENT_LISTEN_ADDR", "")
		t.Setenv("NETZBREMSE_IMPORT_DIR", "")
		t.Setenv("NETZBREMSE_ENDPOINT", "")
		t.Setenv("NETZBREMSE_SPEEDTEST_COMMAND", "")
		t.Setenv("NETZBREMSE_SPEEDTEST_TIMEOUT", "")

		_, err := LoadMeasurement()
		if err == nil {
			t.Fatal("expected error for invalid interval")
		}
	})

	t.Run("custom values", func(t *testing.T) {
		t.Setenv("NETZBREMSE_MEASUREMENT_INTERVAL", "15m")
		t.Setenv("NETZBREMSE_MEASUREMENT_LISTEN_ADDR", ":9090")
		t.Setenv("NETZBREMSE_IMPORT_DIR", "/data/import")
		t.Setenv("NETZBREMSE_ENDPOINT", "https://custom.example.com")
		t.Setenv("NETZBREMSE_SPEEDTEST_COMMAND", "/usr/bin/speedtest --json")
		t.Setenv("NETZBREMSE_SPEEDTEST_TIMEOUT", "5m")

		cfg, err := LoadMeasurement()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.ListenAddress != ":9090" {
			t.Errorf("ListenAddress = %q, want %q", cfg.ListenAddress, ":9090")
		}
		if cfg.ImportDir != "/data/import" {
			t.Errorf("ImportDir = %q, want %q", cfg.ImportDir, "/data/import")
		}
		if cfg.Interval != 15*time.Minute {
			t.Errorf("Interval = %v, want %v", cfg.Interval, 15*time.Minute)
		}
		if cfg.Endpoint != "https://custom.example.com" {
			t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "https://custom.example.com")
		}
		if cfg.Command != "/usr/bin/speedtest --json" {
			t.Errorf("Command = %q, want %q", cfg.Command, "/usr/bin/speedtest --json")
		}
		if cfg.Timeout != 5*time.Minute {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 5*time.Minute)
		}
	})
}

func TestLoadDashboard(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DASHBOARD_LISTEN_ADDR", "")
		t.Setenv("NETZBREMSE_DASHBOARD_LIMIT", "")

		cfg, err := LoadDashboard()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ListenAddress != ":8501" {
			t.Errorf("ListenAddress = %q, want %q", cfg.ListenAddress, ":8501")
		}
		if cfg.Limit != 50 {
			t.Errorf("Limit = %d, want %d", cfg.Limit, 50)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DASHBOARD_LISTEN_ADDR", ":3000")
		t.Setenv("NETZBREMSE_DASHBOARD_LIMIT", "100")

		cfg, err := LoadDashboard()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ListenAddress != ":3000" {
			t.Errorf("ListenAddress = %q, want %q", cfg.ListenAddress, ":3000")
		}
		if cfg.Limit != 100 {
			t.Errorf("Limit = %d, want %d", cfg.Limit, 100)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DASHBOARD_LISTEN_ADDR", "")
		t.Setenv("NETZBREMSE_DASHBOARD_LIMIT", "not-a-number")

		_, err := LoadDashboard()
		if err == nil {
			t.Fatal("expected error for invalid limit")
		}
	})

	t.Run("zero limit rejected", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DASHBOARD_LISTEN_ADDR", "")
		t.Setenv("NETZBREMSE_DASHBOARD_LIMIT", "0")

		_, err := LoadDashboard()
		if err == nil {
			t.Fatal("expected error for zero limit")
		}
	})

	t.Run("negative limit rejected", func(t *testing.T) {
		t.Setenv("NETZBREMSE_DASHBOARD_LISTEN_ADDR", "")
		t.Setenv("NETZBREMSE_DASHBOARD_LIMIT", "-5")

		_, err := LoadDashboard()
		if err == nil {
			t.Fatal("expected error for negative limit")
		}
	})
}

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		fallback string
		want     string
	}{
		{
			name:     "uses env value when set",
			envValue: "custom-value",
			fallback: "default-value",
			want:     "custom-value",
		},
		{
			name:     "uses fallback when env empty",
			envValue: "",
			fallback: "default-value",
			want:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "NETZBREMSE_TEST_ENV_OR_DEFAULT_" + tt.name
			t.Setenv(key, tt.envValue)

			got := envOrDefault(key, tt.fallback)
			if got != tt.want {
				t.Errorf("envOrDefault(%q, %q) = %q, want %q", key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestDurationOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		fallback time.Duration
		want     time.Duration
	}{
		{
			name:     "uses env value when valid",
			envValue: "45s",
			fallback: 1 * time.Minute,
			want:     45 * time.Second,
		},
		{
			name:     "uses fallback when env empty",
			envValue: "",
			fallback: 1 * time.Minute,
			want:     1 * time.Minute,
		},
		{
			name:     "uses fallback when env invalid",
			envValue: "bad-duration",
			fallback: 2 * time.Minute,
			want:     2 * time.Minute,
		},
		{
			name:     "parses complex duration",
			envValue: "1h30m",
			fallback: 1 * time.Hour,
			want:     90 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "NETZBREMSE_TEST_DURATION_" + tt.name
			t.Setenv(key, tt.envValue)

			got := durationOrDefault(key, tt.fallback)
			if got != tt.want {
				t.Errorf("durationOrDefault(%q, %v) = %v, want %v", key, tt.fallback, got, tt.want)
			}
		})
	}
}
