package internal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
)

func RemoveExpiredTorrents(delugeClient DelugeClient, discordClient DiscordClient, config AppConfig) {
	finishedTorrents, err := delugeClient.GetFinishedTorrents()
	if err != nil {
		log.Error("cannot get list of finished torrents", "err", err)
		return
	}
	torrentsToRemove := make([]Torrent, 0)
	for _, torrent := range finishedTorrents {
		if torrent.SecondsSinceDownload > config.RetentionInSeconds() {
			torrentsToRemove = append(torrentsToRemove, torrent)
		}
	}
	if len(torrentsToRemove) == 0 {
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

	formattedTorrents := formatTorrentsToPrint(torrentsToRemove)
	log.Info(
		"remove expired torrents",
		"total released space", FormattedBytes(totalBytes),
		"removed torrents", formattedTorrents,
	)
	sendInfoAboutTorrentsToDiscord(discordClient, torrentsToRemove, totalBytes, formattedTorrents)
}

func formatTorrentsToPrint(torrents []Torrent) string {
	formated := make([]string, 0)
	for i, torrent := range torrents {
		formated = append(formated, fmt.Sprintf("%d. %s (%s)", i+1, torrent.Name, FormattedBytes(torrent.TotalSizeInBytes)))
	}
	return strings.Join(formated, "\n")
}

func sendInfoAboutTorrentsToDiscord(
	discordClient DiscordClient,
	torrentsToRemove []Torrent,
	totalBytes int64,
	listOfTorrents string,
) {
	discordClient.SendInfo(
		fmt.Sprintf(
			`### Retention Job
Removed torrents: %d
Total released space: %s
List of torrents: 
%s`,
			len(torrentsToRemove),
			FormattedBytes(totalBytes),
			listOfTorrents,
		),
	)
}
