package internal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/dustin/go-humanize"
)

func RemoveExpiredTorrents(ctx *AppContext) {
	if ctx.AppConfig.DiskFreeSpaceThreshold != nil {
		usageThreshold := ctx.AppConfig.DiskFreeSpaceThresholdInBytes()
		freeSpaceOfDisk, err := ctx.AppConfig.FreeSpaceOfDiskPathInBytes()
		if err != nil {
			log.Error("cannot check disk stats", "err", err)
			return
		}
		diskFreeSpaceHumanFormat := humanize.Bytes(*freeSpaceOfDisk)
		thresholdHumanFormat := humanize.Bytes(*usageThreshold)
		if *freeSpaceOfDisk > *usageThreshold {
			log.Info("skipping retention, threshold not excedded", "free space", diskFreeSpaceHumanFormat, "threshold", thresholdHumanFormat)
			return
		}
		exceededByHumanFormat := humanize.Bytes(*usageThreshold - *freeSpaceOfDisk)
		log.Info("threshold excedded", "threshold", thresholdHumanFormat, "free space", diskFreeSpaceHumanFormat, "exceeded by", exceededByHumanFormat)
	}
	finishedTorrents, err := ctx.DelugeClient.GetFinishedTorrents()
	if err != nil {
		log.Error("cannot get list of finished torrents", "err", err)
		return
	}
	torrentsToRemove := make([]Torrent, 0)
	for _, torrent := range finishedTorrents {
		if torrent.SecondsSinceDownload > ctx.AppConfig.RetentionInSeconds() {
			torrentsToRemove = append(torrentsToRemove, torrent)
		}
	}
	if len(torrentsToRemove) == 0 {
		log.Info("no expired torrents to remove")
		return
	}
	err = ctx.DelugeClient.RemoveTorrentsWithData(torrentsToRemove, ctx.AppConfig.DryRun)
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
	sendInfoAboutTorrentsToDiscord(ctx.DiscordClient, torrentsToRemove, totalBytes, formattedTorrents)
}

func formatTorrentsToPrint(torrents []Torrent) string {
	formated := make([]string, 0)
	for i, torrent := range torrents {
		formated = append(formated, fmt.Sprintf("%d. %s (%s)", i+1, torrent.Name, FormattedBytes(torrent.TotalSizeInBytes)))
	}
	return strings.Join(formated, "\n")
}

func sendInfoAboutTorrentsToDiscord(
	discordClient *DiscordClient,
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
