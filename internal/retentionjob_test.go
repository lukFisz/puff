package internal

import (
	"testing"
)

func TestRemoveExpiredTorrents_NoExpiredTorrents(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		TorrentDeluge{
			Hash:                 "abc123",
			TorrentName:          "Recent.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 60, // 1 minute old
		},
	}

	config := AppConfig{
		Retention:              "P14D", // 14 days
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		TorrentClientTimeout:   "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	RemoveExpiredTorrents(ctx, mockClient)

	if !mockClient.GetFinishedTorrentsCalled {
		t.Error("GetFinishedTorrents should have been called")
	}

	if mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should not have been called for non-expired torrents")
	}

	// Clean up scheduler
	scheduler.Shutdown()
}

func TestRemoveExpiredTorrents_WithExpiredTorrents(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		TorrentDeluge{
			Hash:                 "abc123",
			TorrentName:          "Old.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 15 * 24 * 60 * 60, // 15 days old
		},
		TorrentDeluge{
			Hash:                 "def456",
			TorrentName:          "Very.Old.Torrent",
			TotalSizeInBytes:     500 * 1024 * 1024,
			SecondsSinceDownload: 30 * 24 * 60 * 60, // 30 days old
		},
		TorrentDeluge{
			Hash:                 "ghi789",
			TorrentName:          "Recent.Torrent",
			TotalSizeInBytes:     200 * 1024 * 1024,
			SecondsSinceDownload: 5 * 24 * 60 * 60, // 5 days old
		},
	}

	config := AppConfig{
		Retention:              "P14D", // 14 days
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		TorrentClientTimeout:   "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	RemoveExpiredTorrents(ctx, mockClient)

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
		if !expectedNames[torrent.Name()] {
			t.Errorf("Unexpected torrent marked for removal: %s", torrent.Name())
		}
	}

	scheduler.Shutdown()
}

func TestRemoveExpiredTorrents_EmptyTorrentList(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{}

	config := AppConfig{
		Retention:              "P14D",
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		TorrentClientTimeout:   "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	RemoveExpiredTorrents(ctx, mockClient)

	if !mockClient.GetFinishedTorrentsCalled {
		t.Error("GetFinishedTorrents should have been called")
	}

	if mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should not have been called for empty torrent list")
	}

	scheduler.Shutdown()
}

func TestRemoveExpiredTorrents_DryRunMode(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		TorrentDeluge{
			Hash:                 "abc123",
			TorrentName:          "Old.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 20 * 24 * 60 * 60, // 20 days old
		},
	}

	config := AppConfig{
		Retention:              "P14D",
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		TorrentClientTimeout:   "2m",
		DryRun:                 true, // DRY RUN MODE
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	RemoveExpiredTorrents(ctx, mockClient)

	if !mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should have been called even in dry run")
	}

	if !mockClient.RemoveTorrentsWithDataDryRun {
		t.Error("DryRun flag should have been passed as true")
	}

	scheduler.Shutdown()
}

func TestRemoveExpiredTorrents_GetTorrentsError(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsError = errTest

	config := AppConfig{
		Retention:              "P14D",
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		TorrentClientTimeout:   "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	// Should not panic, just log error and return
	RemoveExpiredTorrents(ctx, mockClient)

	if mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should not be called when GetFinishedTorrents fails")
	}

	scheduler.Shutdown()
}

func TestRemoveExpiredTorrents_RemovalError(t *testing.T) {
	mockClient := NewMockTorrentClient()
	mockClient.GetFinishedTorrentsResult = []Torrent{
		TorrentDeluge{
			Hash:                 "abc123",
			TorrentName:          "Old.Torrent",
			TotalSizeInBytes:     100 * 1024 * 1024,
			SecondsSinceDownload: 20 * 24 * 60 * 60,
		},
	}
	mockClient.RemoveTorrentsWithDataError = errTest

	config := AppConfig{
		Retention:              "P14D",
		DelugeUrl:              "http://test",
		DelugePassword:         "test",
		TorrentClientTimeout:   "2m",
		DryRun:                 false,
		DiscordWebhookUrl:      "",
		DiskFreeSpaceThreshold: nil,
	}

	discordClient := NewDiscordClient(config)
	scheduler := NewScheduler(config)
	shutdownChan := make(chan bool)

	ctx := &AppContext{
		Scheduler:     &scheduler,
		DiscordClient: discordClient,
		AppConfig:     &config,
		ShutdownChan:  &shutdownChan,
	}

	// Should not panic, just log error
	RemoveExpiredTorrents(ctx, mockClient)

	if !mockClient.RemoveTorrentsWithDataCalled {
		t.Error("RemoveTorrentsWithData should have been called")
	}

	scheduler.Shutdown()
}

func TestNewAppContext(t *testing.T) {
	config := AppConfig{
		Cron:                 "0 0 * * *",
		DelugeUrl:            "http://test",
		DelugePassword:       "test",
		TorrentClientTimeout: "2m",
		Preview:              true,
		DiscordWebhookUrl:    "",
		TimeZone:             "UTC",
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

	if ctx.Jobs == nil {
		t.Error("Jobs should not be nil")
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
		Cron:                 "0 0 * * *",
		DelugeUrl:            "http://test",
		DelugePassword:       "test",
		TorrentClientTimeout: "2m",
		Preview:              true,
		DiscordWebhookUrl:    "",
		TimeZone:             "UTC",
	}

	ctx := NewAppContext(config)

	if len(*ctx.Jobs) != 1 {
		t.Fatalf("Expected 1 job in preview mode, got %d", len(*ctx.Jobs))
	}

	job := (*ctx.Jobs)[0]
	if job.TorrentType != "preview" {
		t.Errorf("Expected TorrentType 'preview', got '%s'", job.TorrentType)
	}

	// Clean up
	(*ctx.Scheduler).Shutdown()
}

func TestNewAppContext_DelugeClient(t *testing.T) {
	config := AppConfig{
		Cron:                 "0 0 * * *",
		DelugeUrl:            "http://test",
		DelugePassword:       "test",
		TorrentClientTimeout: "2m",
		Preview:              false,
		DiscordWebhookUrl:    "",
		TimeZone:             "UTC",
	}

	ctx := NewAppContext(config)

	if len(*ctx.Jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(*ctx.Jobs))
	}

	job := (*ctx.Jobs)[0]
	if job.TorrentType != "deluge" {
		t.Errorf("Expected TorrentType 'deluge', got '%s'", job.TorrentType)
	}

	_, isDeluge := job.TorrentClient.(*DelugeClient)
	if !isDeluge {
		t.Error("Expected DelugeClient when DelugeUrl is set")
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
