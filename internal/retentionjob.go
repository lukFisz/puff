package internal

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
)

func RemoveExpiredTorrents(ctx *AppContext, torrentClient TorrentClient) {
	if ctx.AppConfig.DiskFreeSpaceThreshold != nil {
		usageThreshold := ctx.AppConfig.DiskFreeSpaceThresholdInBytes()
		freeSpaceOfDisk, err := ctx.AppConfig.FreeSpaceOfDiskPathInBytes()
		if err != nil {
			log.Error().Err(err).Msg("cannot check disk stats")
			return
		}
		diskFreeSpaceHumanFormat := humanize.Bytes(*freeSpaceOfDisk)
		thresholdHumanFormat := humanize.Bytes(*usageThreshold)
		if *freeSpaceOfDisk > *usageThreshold {
			log.Info().Str("free space", diskFreeSpaceHumanFormat).Str("threshold", thresholdHumanFormat).Msg("skipping retention, threshold not excedded")
			return
		}
		exceededByHumanFormat := humanize.Bytes(*usageThreshold - *freeSpaceOfDisk)
		log.Info().Str("threshold", thresholdHumanFormat).Str("free space", diskFreeSpaceHumanFormat).Str("exceeded by", exceededByHumanFormat).Msg("threshold excedded")
	}
	finishedTorrents, err := torrentClient.GetFinishedTorrents()
	if err != nil {
		log.Error().Err(err).Msg("cannot get list of finished torrents")
		return
	}
	torrentsToRemove := make([]Torrent, 0)
	for _, torrent := range finishedTorrents {
		if torrent.Expired(ctx.AppConfig.RetentionInSeconds()) {
			torrentsToRemove = append(torrentsToRemove, torrent)
		}
	}
	if len(torrentsToRemove) == 0 {
		log.Info().Msg("no expired torrents to remove")
		return
	}
	err = torrentClient.RemoveTorrentsWithData(torrentsToRemove, ctx.AppConfig.DryRun)
	if err != nil {
		log.Error().Err(err).Msg("remove torrents failed")
		return
	}

	var totalBytes int64 = 0
	for _, torrent := range torrentsToRemove {
		totalBytes += torrent.SizeInBytes()
	}

	formattedTorrents := formatTorrentsToPrint(torrentsToRemove)
	log.Info().
		Str("total released space", FormattedBytes(totalBytes)).
		Str("removed torrents", formattedTorrents).
		Msg("remove expired torrents")
	sendInfoAboutTorrentsToDiscord(ctx.DiscordClient, torrentsToRemove, totalBytes, formattedTorrents)
}

func formatTorrentsToPrint(torrents []Torrent) string {
	formated := make([]string, 0)
	for i, torrent := range torrents {
		formated = append(formated, fmt.Sprintf("%d. %s (%s)", i+1, torrent.Name(), FormattedBytes(torrent.SizeInBytes())))
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
