package internal

import (
	"testing"
)

func TestFormattedBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 500,
			want:  "500 B",
		},
		{
			name:  "kilobytes",
			bytes: 5 * 1024,
			want:  "5.00 KB",
		},
		{
			name:  "megabytes",
			bytes: 100 * 1024 * 1024,
			want:  "100.00 MB",
		},
		{
			name:  "gigabytes",
			bytes: 50 * 1024 * 1024 * 1024,
			want:  "50.00 GB",
		},
		{
			name:  "terabytes",
			bytes: 2 * 1024 * 1024 * 1024 * 1024,
			want:  "2.00 TB",
		},
		{
			name:  "zero bytes",
			bytes: 0,
			want:  "0 B",
		},
		{
			name:  "fractional MB",
			bytes: 150*1024*1024 + 512*1024,
			want:  "150.50 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormattedBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormattedBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestNewDelugeClient(t *testing.T) {
	cfg := AppConfig{
		DelugeUrl:           "http://deluge.test/json",
		DelugePassword:      "testpass",
		DelugeClientTimeout: "30s",
	}

	client := NewDelugeClient(cfg)

	if client.BaseURL != cfg.DelugeUrl {
		t.Errorf("BaseURL = %v, want %v", client.BaseURL, cfg.DelugeUrl)
	}

	if client.Password != cfg.DelugePassword {
		t.Errorf("Password = %v, want %v", client.Password, cfg.DelugePassword)
	}

	if client.Client == nil {
		t.Error("Client should not be nil")
	}

	if client.idCount != 0 {
		t.Errorf("idCount = %v, want %v", client.idCount, 0)
	}

	timeout := client.Client.Timeout
	expected := cfg.DelugeClientTimeoutDuration()
	if timeout != expected {
		t.Errorf("Client.Timeout = %v, want %v", timeout, expected)
	}
}
