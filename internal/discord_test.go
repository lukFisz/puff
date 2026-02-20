package internal

import (
	"bytes"
	"testing"
)

func TestNewDiscordClient(t *testing.T) {
	tests := []struct {
		name       string
		webhookUrl string
		wantNil    bool
	}{
		{
			name:       "with webhook URL",
			webhookUrl: "https://discord.com/api/webhooks/123/abc",
			wantNil:    false,
		},
		{
			name:       "without webhook URL",
			webhookUrl: "",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AppConfig{DiscordWebhookUrl: tt.webhookUrl}
			client := NewDiscordClient(config)

			if client == nil {
				t.Error("NewDiscordClient() returned nil, expected client")
				return
			}

			if tt.wantNil {
				if client.webhookUrl != nil {
					t.Errorf("webhookUrl = %v, want nil", client.webhookUrl)
				}
			} else {
				if client.webhookUrl == nil {
					t.Error("webhookUrl is nil, expected value")
				} else if *client.webhookUrl != tt.webhookUrl {
					t.Errorf("webhookUrl = %v, want %v", *client.webhookUrl, tt.webhookUrl)
				}
			}
		})
	}
}

func TestDiscordClient_Write(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		shouldMatch bool
	}{
		{
			name:        "error level log",
			input:       []byte(`{"level":"error","message":"something went wrong"}`),
			shouldMatch: true,
		},
		{
			name:        "fatal level log",
			input:       []byte(`{"level":"fatal","message":"critical failure"}`),
			shouldMatch: true,
		},
		{
			name:        "panic level log",
			input:       []byte(`{"level":"panic","message":"panic"}`),
			shouldMatch: true,
		},
		{
			name:        "info level log",
			input:       []byte(`{"level":"info","message":"everything is fine"}`),
			shouldMatch: false,
		},
		{
			name:        "debug level log",
			input:       []byte(`{"level":"debug","message":"detailed info"}`),
			shouldMatch: false,
		},
		{
			name:        "warn level log",
			input:       []byte(`{"level":"warn","message":"warning"}`),
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})

			n, err := client.Write(tt.input)

			if err != nil {
				t.Errorf("Write() error = %v, want nil", err)
			}

			if n != len(tt.input) {
				t.Errorf("Write() returned %v bytes, want %v", n, len(tt.input))
			}
		})
	}
}

func TestDiscordClient_WriteInvalidJSON(t *testing.T) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})

	_, err := client.Write([]byte("not json"))

	if err == nil {
		t.Error("Write() expected error for invalid JSON, got nil")
	}
}

func TestFormatTorrentsToPrint(t *testing.T) {
	tests := []struct {
		name     string
		torrents []Torrent
		want     string
	}{
		{
			name:     "empty list",
			torrents: []Torrent{},
			want:     "",
		},
		{
			name: "single torrent",
			torrents: []Torrent{
				TorrentDeluge{
					Hash:                 "abc123",
					TorrentName:          "Test.Torrent",
					TotalSizeInBytes:     100 * 1024 * 1024,
					SecondsSinceDownload: 3600,
				},
			},
			want: "1. Test.Torrent (100.00 MB)",
		},
		{
			name: "multiple torrents",
			torrents: []Torrent{
				TorrentDeluge{
					Hash:                 "abc123",
					TorrentName:          "First.Torrent",
					TotalSizeInBytes:     100 * 1024 * 1024,
					SecondsSinceDownload: 3600,
				},
				TorrentDeluge{
					Hash:                 "def456",
					TorrentName:          "Second.Torrent",
					TotalSizeInBytes:     500 * 1024 * 1024,
					SecondsSinceDownload: 7200,
				},
			},
			want: "1. First.Torrent (100.00 MB)\n2. Second.Torrent (500.00 MB)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTorrentsToPrint(tt.torrents)
			if got != tt.want {
				t.Errorf("formatTorrentsToPrint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendInfoAboutTorrentsToDiscord(t *testing.T) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})

	torrents := []Torrent{
		TorrentDeluge{
			Hash:                 "abc123",
			TorrentName:          "Test.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 3600,
		},
	}

	totalBytes := int64(100 * 1024 * 1024)
	listOfTorrents := "1. Test.Torrent (100.00 MB)"

	// Should not panic
	sendInfoAboutTorrentsToDiscord(client, torrents, totalBytes, listOfTorrents)
}

func TestDiscordClient_SendInfo(t *testing.T) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})
	client.SendInfo("Test message")
}

func TestDiscordClient_SendError(t *testing.T) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})
	client.SendError("Test error message")
}

func BenchmarkDiscordClient_Write(b *testing.B) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})
	data := []byte(`{"level":"error","message":"benchmark test"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Write(data)
	}
}

func BenchmarkFormatTorrentsToPrint(b *testing.B) {
	torrents := make([]Torrent, 100)
	for i := 0; i < 100; i++ {
		torrents[i] = TorrentDeluge{
			Hash:                 "hash" + string(rune(i)),
			TorrentName:          "Torrent.Name." + string(rune(i)),
			TotalSizeInBytes:     int64(i) * 1024 * 1024,
			SecondsSinceDownload: i * 3600,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatTorrentsToPrint(torrents)
	}
}

func TestDiscordClientWriteImplementsIOWriter(t *testing.T) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})

	var buf bytes.Buffer
	testData := []byte(`{"level":"info","message":"test"}`)

	n1, err1 := buf.Write(testData)
	n2, err2 := client.Write(testData)

	if err1 != nil || err2 != nil {
		t.Errorf("Write errors: buf=%v, client=%v", err1, err2)
	}

	if n1 != n2 {
		t.Errorf("Write counts differ: buf=%d, client=%d", n1, n2)
	}
}
