package internal

import (
	"testing"
	"time"
)

func TestNewQbitTorrentClient(t *testing.T) {
	cfg := AppConfig{
		QbitTorrentUrl:       "http://qbit.test:8080",
		QbitTorrentUsername:  "admin",
		QbitTorrentPassword:  "testpass",
		TorrentClientTimeout: "30s",
	}

	client := NewQbitTorrentClient(cfg)

	if client.BaseURL != cfg.QbitTorrentUrl {
		t.Errorf("BaseURL = %v, want %v", client.BaseURL, cfg.QbitTorrentUrl)
	}

	if client.Username != cfg.QbitTorrentUsername {
		t.Errorf("Username = %v, want %v", client.Username, cfg.QbitTorrentUsername)
	}

	if client.Password != cfg.QbitTorrentPassword {
		t.Errorf("Password = %v, want %v", client.Password, cfg.QbitTorrentPassword)
	}

	if client.Client == nil {
		t.Error("Client should not be nil")
	}
}

func TestTorrentQbit_Expired(t *testing.T) {
	tests := []struct {
		name         string
		completionOn int64
		maxSec       int
		wantExpired  bool
	}{
		{
			name:         "expired torrent",
			completionOn: time.Now().Add(-15 * 24 * time.Hour).Unix(),
			maxSec:       14 * 24 * 60 * 60,
			wantExpired:  true,
		},
		{
			name:         "not expired torrent",
			completionOn: time.Now().Add(-5 * 24 * time.Hour).Unix(),
			maxSec:       14 * 24 * 60 * 60,
			wantExpired:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			torrent := TorrentQbit{
				Hash:             "abc123",
				TorrentName:      "Test.Torrent",
				TotalSizeInBytes: 100 * 1024 * 1024,
				CompletionOn:     tt.completionOn,
			}
			got := torrent.Expired(tt.maxSec)
			if got != tt.wantExpired {
				t.Errorf("Expired(%d) = %v, want %v", tt.maxSec, got, tt.wantExpired)
			}
		})
	}
}

func TestTorrentQbit_InterfaceMethods(t *testing.T) {
	torrent := TorrentQbit{
		Hash:             "def456",
		TorrentName:      "Qbit.Torrent",
		TotalSizeInBytes: 500 * 1024 * 1024,
		CompletionOn:     time.Now().Unix(),
	}

	if torrent.Id() != "def456" {
		t.Errorf("Id() = %v, want %v", torrent.Id(), "def456")
	}

	if torrent.Name() != "Qbit.Torrent" {
		t.Errorf("Name() = %v, want %v", torrent.Name(), "Qbit.Torrent")
	}

	if torrent.SizeInBytes() != 500*1024*1024 {
		t.Errorf("SizeInBytes() = %v, want %v", torrent.SizeInBytes(), 500*1024*1024)
	}

	// Verify TorrentQbit implements Torrent interface
	var _ Torrent = torrent
}

func TestTorrentQbit_CompletionTime(t *testing.T) {
	now := time.Now().Unix()
	torrent := TorrentQbit{
		CompletionOn: now,
	}

	completionTime := torrent.CompletionTime()
	if completionTime.Unix() != now {
		t.Errorf("CompletionTime().Unix() = %v, want %v", completionTime.Unix(), now)
	}
}
