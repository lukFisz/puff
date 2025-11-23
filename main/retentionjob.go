package main

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
)

func NewRemoveExpiredTorrentsJob(config AppConfig) func() {
	delugeClient := NewDelugeClient(config.DelugeUrl, config.DelugePassword, config.DelugeClientTimeoutDuration())
	if err := delugeClient.Login(); err != nil {
		log.Error("deluge client", "err", err)
	} else {
		log.Info("deluge client", "authentication", "successful")
	}
	discordClient := NewDiscordClient(config)
	return func() {
		RemoveExpiredTorrents(*delugeClient, discordClient, config)
	}
}

func RemoveExpiredTorrents(delugeClient DelugeClient, discordClient *DiscordClient, config AppConfig) {
	finishedTorrents, err := delugeClient.GetFinishedTorrents()
	if err != nil {
		log.Error("cannot get list of finished torrents", "err", err)
	}
	torrentsToRemove := make([]Torrent, 0)
	for _, torrent := range finishedTorrents {
		if torrent.SecondsSinceDownload > config.RetentionInSeconds() {
			torrentsToRemove = append(torrentsToRemove, torrent)
		}
	}
	if err == nil && len(torrentsToRemove) == 0 {
		log.Info("no expired torrents to remove")
		return
	}
	err = delugeClient.RemoveTorrentsWithData(torrentsToRemove, config.DryRun)
	if err != nil {
		log.Error("remove torrents failed", "err", err)
		return
	}

	var totalBytes int64 = 0
	for _, torrent := range torrentsToRemove {
		totalBytes += torrent.TotalSizeInBytes
	}
	prettyList, err := json.MarshalIndent(torrentsToRemove, "", "  ")
	if err != nil {
		log.Info(
			"remove expired torrents",
			"total released space", FormattedBytes(totalBytes),
			"removed torrents", torrentsToRemove,
		)
		sendInfoAboutTorrentsToDiscord(discordClient, torrentsToRemove, totalBytes, torrentsToRemove)
	} else {
		strPrettyList := string(prettyList)
		log.Info(
			"remove expired torrents",
			"total released space", FormattedBytes(totalBytes),
			"removed torrents", strPrettyList,
		)
		sendInfoAboutTorrentsToDiscord(discordClient, torrentsToRemove, totalBytes, strPrettyList)
	}
}

func sendInfoAboutTorrentsToDiscord(
	discordClient *DiscordClient,
	torrentsToRemove []Torrent,
	totalBytes int64,
	listOfTorrents interface{},
) {
	discordClient.SendInfo(
		fmt.Sprintf(`
### Retention Job
Removed torrents: %d
Total released space: %s
Expired torrents: %s
`, len(torrentsToRemove), FormattedBytes(totalBytes), listOfTorrents),
	)
}
