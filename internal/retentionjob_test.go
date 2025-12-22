package internal

import (
	"testing"
)

func TestRemoveExpiredTorrents_NoExpiredTorrents(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		{
			Hash:                 "abc123",
			Name:                 "Recent.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 60, // 1 minute old
		},
	}

	config := AppConfig{
		Retention:              "P14D", // 14 days
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		DelugeClientTimeout:    "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	var client TorrentClient = mockClient
	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		TorrentClient: &client,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	RemoveExpiredTorrents(ctx)

	if !mockClient.GetFinishedTorrentsCalled {
		t.Error("GetFinishedTorrents should have been called")
	}

	if mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should not have been called for non-expired torrents")
	}
}

func TestRemoveExpiredTorrents_WithExpiredTorrents(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		{
			Hash:                 "abc123",
			Name:                 "Old.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 15 * 24 * 60 * 60, // 15 days old
		},
		{
			Hash:                 "def456",
			Name:                 "Very.Old.Torrent",
			TotalSizeInBytes:     500 * 1024 * 1024,
			SecondsSinceDownload: 30 * 24 * 60 * 60, // 30 days old
		},
		{
			Hash:                 "ghi789",
			Name:                 "Recent.Torrent",
			TotalSizeInBytes:     200 * 1024 * 1024,
			SecondsSinceDownload: 5 * 24 * 60 * 60, // 5 days old
		},
	}

	config := AppConfig{
		Retention:              "P14D", // 14 days
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		DelugeClientTimeout:    "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	var client TorrentClient = mockClient
	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		TorrentClient: &client,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	RemoveExpiredTorrents(ctx)

	if !mockClient.GetFinishedTorrentsCalled {
		t.Error("GetFinishedTorrents should have been called")
	}

	if !mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should have been called")
	}

	if len(mockClient.RemoveTorrentsWithDataInput) != 2 {
		t.Errorf("Expected 2 torrents to be removed, got %d", len(mockClient.RemoveTorrentsWithDataInput))
	}

	// Verify the correct torrents were selected for removal
	expectedNames := map[string]bool{"Old.Torrent": true, "Very.Old.Torrent": true}
	for _, torrent := range mockClient.RemoveTorrentsWithDataInput {
		if !expectedNames[torrent.Name] {
			t.Errorf("Unexpected torrent marked for removal: %s", torrent.Name)
		}
	}
}

func TestRemoveExpiredTorrents_DryRunMode(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		{
			Hash:                 "abc123",
			Name:                 "Old.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 20 * 24 * 60 * 60, // 20 days old
		},
	}

	config := AppConfig{
		Retention:              "P14D",
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		DelugeClientTimeout:    "2m",
		DryRun:                 true, // DRY RUN MODE
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	var client TorrentClient = mockClient
	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		TorrentClient: &client,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	RemoveExpiredTorrents(ctx)

	if !mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should have been called even in dry run")
	}

	if !mockClient.RemoveTorrentsWithDataDryRun {
		t.Error("DryRun flag should have been passed as true")
	}
}

func TestRemoveExpiredTorrents_GetTorrentsError(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsError = errTest

	config := AppConfig{
		Retention:              "P14D",
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		DelugeClientTimeout:    "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	var client TorrentClient = mockClient
	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		TorrentClient: &client,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	// Should not panic, just log error and return
	RemoveExpiredTorrents(ctx)

	if mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should not be called when GetFinishedTorrents fails")
	}
}

func TestRemoveExpiredTorrents_RemovalError(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		{
			Hash:                 "abc123",
			Name:                 "Old.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 20 * 24 * 60 * 60,
		},
	}
	mockClient.RemoveTorrentsWithDataError = errTest

	config := AppConfig{
		Retention:              "P14D",
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		DelugeClientTimeout:    "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	var client TorrentClient = mockClient
	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		TorrentClient: &client,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	// Should not panic, just log error
	RemoveExpiredTorrents(ctx)

	if !mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should have been called")
	}
}

func TestNewAppContext(t *testing.T) {
	config := AppConfig{
		Cron:                "0 0 * * *",
		DelugeUrl:           "http://test",
		DelugePassword:      "test",
		DelugeClientTimeout: "2m",
		Preview:             true,
		DiscordWebhookUrl:   "",
		TimeZone:            "UTC",
	}

	ctx := NewAppContext(config)

	if ctx == nil {
		t.Fatal("NewAppContext returned nil")
	}

	if ctx.Scheduler == nil {
		t.Error("Scheduler should not be nil")
	}

	if ctx.DiscordClient == nil {
		t.Error("DiscordClient should not be nil")
	}

	if ctx.TorrentClient == nil {
		t.Error("TorrentClient should not be nil")
	}

	if ctx.AppConfig == nil {
		t.Error("AppConfig should not be nil")
	}

	if ctx.ShutdownChan == nil {
		t.Error("ShutdownChan should not be nil")
	}

	// Clean up scheduler
	(*ctx.Scheduler).Shutdown()
}

func TestNewAppContext_PreviewMode(t *testing.T) {
	config := AppConfig{
		Cron:                "0 0 * * *",
		DelugeUrl:           "http://test",
		DelugePassword:      "test",
		DelugeClientTimeout: "2m",
		Preview:             true,
		DiscordWebhookUrl:   "",
		TimeZone:            "UTC",
	}

	ctx := NewAppContext(config)

	// Verify preview client is used
	_, isPreview := (*ctx.TorrentClient).(*PreviewTorrentClient)
	if !isPreview {
		t.Error("Expected PreviewTorrentClient when Preview mode is enabled")
	}

	// Clean up
	(*ctx.Scheduler).Shutdown()
}

func TestNewAppContext_RealClient(t *testing.T) {
	config := AppConfig{
		Cron:                "0 0 * * *",
		DelugeUrl:           "http://test",
		DelugePassword:      "test",
		DelugeClientTimeout: "2m",
		Preview:             false,
		DiscordWebhookUrl:   "",
		TimeZone:            "UTC",
	}

	ctx := NewAppContext(config)

	// Verify deluge client is used
	_, isDeluge := (*ctx.TorrentClient).(*DelugeClient)
	if !isDeluge {
		t.Error("Expected DelugeClient when Preview mode is disabled")
	}

	// Clean up
	(*ctx.Scheduler).Shutdown()
}

// Helper error for tests
var errTest = &testError{"test error"}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
