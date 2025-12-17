package internal

import (
	"bytes"
	"regexp"
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
			name:        "ERROR log",
			input:       []byte("2025/12/17 ERROR something went wrong"),
			shouldMatch: true,
		},
		{
			name:        "FATAL log",
			input:       []byte("2025/12/17 FATAL critical failure"),
			shouldMatch: true,
		},
		{
			name:        "INFO log",
			input:       []byte("2025/12/17 INFO everything is fine"),
			shouldMatch: false,
		},
		{
			name:        "DEBUG log",
			input:       []byte("2025/12/17 DEBUG detailed info"),
			shouldMatch: false,
		},
		{
			name:        "lowercase error",
			input:       []byte("2025/12/17 error something wrong"),
			shouldMatch: true,
		},
		{
			name:        "lowercase fatal",
			input:       []byte("2025/12/17 fatal critical"),
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client without webhook (won't actually send)
			client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})

			n, err := client.Write(tt.input)

			if err != nil {
				t.Errorf("Write() error = %v, want nil", err)
			}

			if n != len(tt.input) {
				t.Errorf("Write() returned %v bytes, want %v", n, len(tt.input))
			}

			// Verify regex matching logic
			matched := regexp.MustCompile(`(?i)(ERRO|FATA)`).Match(tt.input)
			if matched != tt.shouldMatch {
				t.Errorf("Regex match = %v, want %v for input: %s", matched, tt.shouldMatch, tt.input)
			}
		})
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
				{
					Hash:                 "abc123",
					Name:                 "Test.Torrent",
					TotalSizeInBytes:     100 * 1024 * 1024,
					SecondsSinceDownload: 3600,
				},
			},
			want: "1. Test.Torrent (100.00 MB)",
		},
		{
			name: "multiple torrents",
			torrents: []Torrent{
				{
					Hash:                 "abc123",
					Name:                 "First.Torrent",
					TotalSizeInBytes:     100 * 1024 * 1024,
					SecondsSinceDownload: 3600,
				},
				{
					Hash:                 "def456",
					Name:                 "Second.Torrent",
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
	// This test verifies the function doesn't panic
	// Actual sending is skipped when webhookUrl is nil
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})

	torrents := []Torrent{
		{
			Hash:                 "abc123",
			Name:                 "Test.Torrent",
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
	// Test that SendInfo doesn't panic when webhook is not configured
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})
	client.SendInfo("Test message")
	// If we got here without panic, test passed
}

func TestDiscordClient_SendError(t *testing.T) {
	// Test that SendError doesn't panic when webhook is not configured
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})
	client.SendError("Test error message")
	// If we got here without panic, test passed
}

func BenchmarkDiscordClient_Write(b *testing.B) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})
	data := []byte("2025/12/17 ERROR benchmark test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Write(data)
	}
}

func BenchmarkFormatTorrentsToPrint(b *testing.B) {
	torrents := make([]Torrent, 100)
	for i := 0; i < 100; i++ {
		torrents[i] = Torrent{
			Hash:                 "hash" + string(rune(i)),
			Name:                 "Torrent.Name." + string(rune(i)),
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

	// Test that DiscordClient can be used as io.Writer
	var buf bytes.Buffer
	testData := []byte("test data")

	// Write to both buffer and discord client
	n1, err1 := buf.Write(testData)
	n2, err2 := client.Write(testData)

	if err1 != nil || err2 != nil {
		t.Errorf("Write errors: buf=%v, client=%v", err1, err2)
	}

	if n1 != n2 {
		t.Errorf("Write counts differ: buf=%d, client=%d", n1, n2)
	}
}
