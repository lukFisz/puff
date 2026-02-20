package internal

import (
	"testing"
)

func TestNewDelugeClient(t *testing.T) {
	cfg := AppConfig{
		DelugeUrl:            "http://deluge.test/json",
		DelugePassword:       "testpass",
		TorrentClientTimeout: "30s",
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
	expected := cfg.TorrentClientTimeoutDuration()
	if timeout != expected {
		t.Errorf("Client.Timeout = %v, want %v", timeout, expected)
	}
}

func TestTorrentDeluge_Expired(t *testing.T) {
	tests := []struct {
		name                 string
		secondsSinceDownload int
		maxSec               int
		wantExpired          bool
	}{
		{
			name:                 "expired torrent",
			secondsSinceDownload: 15 * 24 * 60 * 60,
			maxSec:               14 * 24 * 60 * 60,
			wantExpired:          true,
		},
		{
			name:                 "not expired torrent",
			secondsSinceDownload: 5 * 24 * 60 * 60,
			maxSec:               14 * 24 * 60 * 60,
			wantExpired:          false,
		},
		{
			name:                 "exactly at threshold",
			secondsSinceDownload: 14 * 24 * 60 * 60,
			maxSec:               14 * 24 * 60 * 60,
			wantExpired:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			torrent := TorrentDeluge{
				Hash:                 "abc123",
				TorrentName:          "Test.Torrent",
				TotalSizeInBytes:     100 * 1024 * 1024,
				SecondsSinceDownload: tt.secondsSinceDownload,
			}
			got := torrent.Expired(tt.maxSec)
			if got != tt.wantExpired {
				t.Errorf("Expired(%d) = %v, want %v", tt.maxSec, got, tt.wantExpired)
			}
		})
	}
}

func TestTorrentDeluge_InterfaceMethods(t *testing.T) {
	torrent := TorrentDeluge{
		Hash:                 "abc123",
		TorrentName:          "Test.Torrent",
		TotalSizeInBytes:     100 * 1024 * 1024,
		SecondsSinceDownload: 60,
	}

	if torrent.Id() != "abc123" {
		t.Errorf("Id() = %v, want %v", torrent.Id(), "abc123")
	}

	if torrent.Name() != "Test.Torrent" {
		t.Errorf("Name() = %v, want %v", torrent.Name(), "Test.Torrent")
	}

	if torrent.SizeInBytes() != 100*1024*1024 {
		t.Errorf("SizeInBytes() = %v, want %v", torrent.SizeInBytes(), 100*1024*1024)
	}

	// Verify TorrentDeluge implements Torrent interface
	var _ Torrent = torrent
}
