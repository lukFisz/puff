package internal

import (
	"os"
	"testing"
	"time"
)

func TestAppConfig_RetentionInSeconds(t *testing.T) {
	tests := []struct {
		name      string
		retention string
		want      int
	}{
		{
			name:      "14 days",
			retention: "P14D",
			want:      14 * 24 * 60 * 60,
		},
		{
			name:      "30 days",
			retention: "P30D",
			want:      30 * 24 * 60 * 60,
		},
		{
			name:      "1 month (approx 30.44 days)",
			retention: "P1M",
			want:      2629746, // Actual approximation used by period library
		},
		{
			name:      "7 days and 12 hours",
			retention: "P7DT12H",
			want:      (7*24 + 12) * 60 * 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AppConfig{Retention: tt.retention}
			got := config.RetentionInSeconds()
			if got != tt.want {
				t.Errorf("RetentionInSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_StartDelayDuration(t *testing.T) {
	tests := []struct {
		name       string
		startDelay string
		want       time.Duration
	}{
		{
			name:       "no delay",
			startDelay: "0s",
			want:       0,
		},
		{
			name:       "30 seconds",
			startDelay: "30s",
			want:       30 * time.Second,
		},
		{
			name:       "2 minutes",
			startDelay: "2m",
			want:       2 * time.Minute,
		},
		{
			name:       "1 hour",
			startDelay: "1h",
			want:       time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AppConfig{StartDelay: tt.startDelay}
			got := config.StartDelayDuration()
			if got != tt.want {
				t.Errorf("StartDelayDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_TorrentClientTimeoutDuration(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
		want    time.Duration
	}{
		{
			name:    "default 2 minutes",
			timeout: "2m0s",
			want:    2 * time.Minute,
		},
		{
			name:    "30 seconds",
			timeout: "30s",
			want:    30 * time.Second,
		},
		{
			name:    "5 minutes",
			timeout: "5m",
			want:    5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AppConfig{TorrentClientTimeout: tt.timeout}
			got := config.TorrentClientTimeoutDuration()
			if got != tt.want {
				t.Errorf("TorrentClientTimeoutDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppConfig_DiskFreeSpaceThresholdInBytes(t *testing.T) {
	tests := []struct {
		name      string
		threshold *string
		want      *uint64
	}{
		{
			name:      "nil threshold",
			threshold: nil,
			want:      nil,
		},
		{
			name:      "100GB",
			threshold: strPtr("100GB"),
			want:      uint64Ptr(100 * 1000 * 1000 * 1000),
		},
		{
			name:      "1TB",
			threshold: strPtr("1TB"),
			want:      uint64Ptr(1000 * 1000 * 1000 * 1000),
		},
		{
			name:      "500MB",
			threshold: strPtr("500MB"),
			want:      uint64Ptr(500 * 1000 * 1000),
		},
		{
			name:      "100GiB",
			threshold: strPtr("100GiB"),
			want:      uint64Ptr(100 * 1024 * 1024 * 1024),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AppConfig{DiskFreeSpaceThreshold: tt.threshold}
			got := config.DiskFreeSpaceThresholdInBytes()
			if got == nil && tt.want == nil {
				return
			}
			if got == nil || tt.want == nil {
				t.Errorf("DiskFreeSpaceThresholdInBytes() = %v, want %v", got, tt.want)
				return
			}
			if *got != *tt.want {
				t.Errorf("DiskFreeSpaceThresholdInBytes() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func TestAppConfig_Location(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		want     string
	}{
		{
			name:     "UTC",
			timezone: "UTC",
			want:     "UTC",
		},
		{
			name:     "Europe/Warsaw",
			timezone: "Europe/Warsaw",
			want:     "Europe/Warsaw",
		},
		{
			name:     "America/New_York",
			timezone: "America/New_York",
			want:     "America/New_York",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AppConfig{TimeZone: tt.timezone}
			got := config.Location()
			if got.String() != tt.want {
				t.Errorf("Location() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"PUFF_DELUGE_URL":      os.Getenv("PUFF_DELUGE_URL"),
		"PUFF_DELUGE_PASSWORD": os.Getenv("PUFF_DELUGE_PASSWORD"),
		"PUFF_PREVIEW_MODE":    os.Getenv("PUFF_PREVIEW_MODE"),
		"PUFF_DRY_RUN":         os.Getenv("PUFF_DRY_RUN"),
	}

	// Restore env vars after test
	defer func() {
		for key, val := range originalVars {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	// Set test env vars
	os.Setenv("PUFF_DELUGE_URL", "http://test.local/json")
	os.Setenv("PUFF_DELUGE_PASSWORD", "testpass")
	os.Setenv("PUFF_PREVIEW_MODE", "true")
	os.Setenv("PUFF_DRY_RUN", "true")

	config := GetConfig()

	if config.DelugeUrl != "http://test.local/json" {
		t.Errorf("DelugeUrl = %v, want %v", config.DelugeUrl, "http://test.local/json")
	}

	if config.DelugePassword != "testpass" {
		t.Errorf("DelugePassword = %v, want %v", config.DelugePassword, "testpass")
	}

	if !config.Preview {
		t.Errorf("Preview = %v, want %v", config.Preview, true)
	}

	if !config.DryRun {
		t.Errorf("DryRun = %v, want %v", config.DryRun, true)
	}

	// Test defaults
	if config.Retention != "P14D" {
		t.Errorf("Retention = %v, want %v (default)", config.Retention, "P14D")
	}

	if config.TorrentClientTimeout != "2m0s" {
		t.Errorf("TorrentClientTimeout = %v, want %v (default)", config.TorrentClientTimeout, "2m0s")
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}
