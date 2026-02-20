package internal

import (
	"github.com/charmbracelet/log"
)

type PreviewTorrentClient struct{}

func NewPreviewTorrentClient() *PreviewTorrentClient {
	return &PreviewTorrentClient{}
}

func (p *PreviewTorrentClient) CheckConnection() {
	log.Info("torrent client", "authentication", "successful")
}

func (p *PreviewTorrentClient) Login() error {
	return nil
}

func (p *PreviewTorrentClient) GetFinishedTorrents() ([]Torrent, error) {
	return []Torrent{
		TorrentDeluge{Hash: "abc123", TorrentName: "Torrent.1", TotalSizeInBytes: 1024 * 1024 * 100, SecondsSinceDownload: 86400 * 30},
		TorrentDeluge{Hash: "def456", TorrentName: "Torrent.2", TotalSizeInBytes: 1024 * 1024 * 500, SecondsSinceDownload: 86400 * 15},
		TorrentDeluge{Hash: "ghi789", TorrentName: "Torrent.3", TotalSizeInBytes: 1024 * 1024 * 1024 * 2, SecondsSinceDownload: 86400 * 7},
	}, nil
}

func (p *PreviewTorrentClient) RemoveTorrentsWithData(torrents []Torrent, dryRun bool) error {
	return nil
}
