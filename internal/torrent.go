package internal

type TorrentClient interface {
	CheckConnection()
	Login() error
	GetFinishedTorrents() ([]Torrent, error)
	RemoveTorrentsWithData(torrents []Torrent, dryRun bool) error
}

type Torrent interface {
	Expired(maxSecToLiveInclusive int) bool
	Id() string
	SizeInBytes() int64
	Name() string
}
