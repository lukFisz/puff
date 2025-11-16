package main

import (
	"encoding/json"

	"github.com/charmbracelet/log"
)

func removeExpiredTorrents(delugeClient DelugeClient, config AppConfig) {
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
	} else {
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
		}
		log.Info(
			"remove expired torrents",
			"total released space", FormattedBytes(totalBytes),
			"removed torrents", string(prettyList),
		)
	}
}
