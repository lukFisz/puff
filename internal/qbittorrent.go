package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"resty.dev/v3"
)

type TorrentQbit struct {
	Hash             string `json:"hash"`
	TotalSizeInBytes int64  `json:"total_size"`
	TorrentName      string `json:"name"`
	CompletionOn     int64  `json:"completion_on"`
}

func (t TorrentQbit) Expired(maxSecToLiveInclusive int) bool {
	return time.Now().Sub(t.CompletionTime()).Seconds() > float64(maxSecToLiveInclusive)
}

func (t TorrentQbit) Id() string {
	return t.Hash
}

func (t TorrentQbit) SizeInBytes() int64 {
	return t.TotalSizeInBytes
}

func (t TorrentQbit) Name() string {
	return t.TorrentName
}

func (t TorrentQbit) CompletionTime() time.Time {
	return time.Unix(t.CompletionOn, 0)
}

type QbitTorrentClient struct {
	BaseURL  string
	Username string
	Password string
	Client   *resty.Client
}

func NewQbitTorrentClient(cfg AppConfig) *QbitTorrentClient {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("qbittorrent init")
	}
	return &QbitTorrentClient{
		BaseURL:  cfg.QbitTorrentUrl,
		Username: cfg.QbitTorrentUsername,
		Password: cfg.QbitTorrentPassword,
		Client: resty.New().
			SetBaseURL(cfg.QbitTorrentUrl).
			SetCookieJar(jar),
	}
}

func (q *QbitTorrentClient) CheckConnection() {
	resp, err := q.Client.R().Post("api/v2/auth/login")
	if err != nil {
		log.Error().Err(err).Str("url", q.BaseURL).Msg("failed to connect to qbittorrent")
	}
	if resp.StatusCode() < 300 || resp.String() == "Fails." {
		log.Info().
			Str("url", q.BaseURL).
			Msg("successfully connected to qbittorrent")
		return
	}
	log.Error().
		Str("response", resp.String()).
		Str("status", resp.Status()).
		Str("url", q.BaseURL).
		Msg("failed to connect to qbittorrent")
}

func (q *QbitTorrentClient) Login() error {
	resp, err := q.Client.R().
		SetFormData(map[string]string{
			"username": q.Username,
			"password": q.Password,
		}).
		Post("api/v2/auth/login")

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusOK && resp.String() == "Ok." {
		log.Info().Str("url", q.BaseURL).Msg("successfully logged in to qbittorrent")
		return nil
	}

	return fmt.Errorf("failed to login to qbittorrent: status(%s) %s", resp.Status(), resp.String())
}

func (q *QbitTorrentClient) GetFinishedTorrents() ([]Torrent, error) {
	resp, err := q.Client.R().Post("api/v2/torrents/info")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusForbidden {
		err := q.Login()
		if err != nil {
			return nil, err
		}
		resp, err = q.Client.R().Post("api/v2/torrents/info")
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("failed to get torrents info: status(%s) %s", resp.Status(), resp.String())
		}
	}

	var torrents []TorrentQbit
	if err := json.Unmarshal(resp.Bytes(), &torrents); err != nil {
		return nil, err
	}

	torrentsToRemove := make([]Torrent, 0)
	for _, torrent := range torrents {
		if torrent.CompletionOn > 0 {
			torrentsToRemove = append(torrentsToRemove, torrent)
		}
	}

	return torrentsToRemove, nil
}

func (q *QbitTorrentClient) RemoveTorrentsWithData(torrents []Torrent, dryRun bool) error {
	if dryRun {
		log.Info().Msg("dry-run mode enabled, not deleting any torrents")
		return nil
	}
	if len(torrents) == 0 {
		log.Warn().Msg("no torrents to remove")
		return nil
	}

	hashes := make([]string, 0, len(torrents))
	for _, torrent := range torrents {
		hashes = append(hashes, torrent.Id())
	}

	resp, err := q.Client.R().
		SetFormData(map[string]string{
			"hashes":      strings.Join(hashes, "|"),
			"deleteFiles": "true",
		}).
		Post("/api/v2/torrents/delete")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
