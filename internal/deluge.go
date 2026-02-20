package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type DelugeClient struct {
	BaseURL  string
	Password string
	Session  string
	Client   *http.Client
	idCount  int
}

type jsonRpcRequest struct {
	Method string `json:"method"`
	Params any    `json:"params"`
	ID     int    `json:"id"`
}

type jsonRpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
	ID     int             `json:"id"`
}

type TorrentDeluge struct {
	Hash                 string `json:"hash"`
	TotalSizeInBytes     int64  `json:"total_size"`
	TorrentName          string `json:"name"`
	SecondsSinceDownload int    `json:"time_since_download"`
}

func (t TorrentDeluge) Expired(maxSecToLiveInclusive int) bool {
	return t.SecondsSinceDownload > maxSecToLiveInclusive
}

func (t TorrentDeluge) Id() string {
	return t.Hash
}

func (t TorrentDeluge) SizeInBytes() int64 {
	return t.TotalSizeInBytes
}

func (t TorrentDeluge) Name() string {
	return t.TorrentName
}

func NewDelugeClient(cfg AppConfig) *DelugeClient {
	return &DelugeClient{
		BaseURL:  cfg.DelugeUrl,
		Password: cfg.DelugePassword,
		Client:   &http.Client{Timeout: cfg.TorrentClientTimeoutDuration()},
		idCount:  0,
	}
}

func (delugeClient *DelugeClient) CheckConnection() {
	if err := delugeClient.Login(); err != nil {
		log.Error().Err(err).Msg("deluge client")
	} else {
		log.Info().Str("authentication", "successful").Msg("deluge client")
	}
}

func (delugeClient *DelugeClient) request(method string, params interface{}, retry401 bool) (json.RawMessage, error) {
	delugeClient.idCount++
	reqData := jsonRpcRequest{
		Method: method,
		Params: params,
		ID:     delugeClient.idCount,
	}
	b, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", delugeClient.BaseURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if delugeClient.Session != "" {
		req.Header.Set("Cookie", "_session_id="+delugeClient.Session)
	}

	resp, err := delugeClient.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp jsonRpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		message := rpcResp.Error.(map[string]any)["message"]
		if retry401 && message != nil && message.(string) == "Not authenticated" {
			if err := delugeClient.Login(); err != nil {
				return nil, err
			}
			return delugeClient.request(method, params, false)
		}
		return nil, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}

	return rpcResp.Result, nil
}

func (delugeClient *DelugeClient) Login() error {
	delugeClient.idCount++
	reqData := jsonRpcRequest{
		Method: "auth.login",
		Params: []interface{}{delugeClient.Password},
		ID:     delugeClient.idCount,
	}
	b, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", delugeClient.BaseURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := delugeClient.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("response status code: %v", resp.StatusCode)
	}

	var rpcResp jsonRpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return err
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error: %v", rpcResp.Error)
	}

	var success bool
	if err := json.Unmarshal(rpcResp.Result, &success); err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("rpc error: authentication failed")
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_session_id" {
			delugeClient.Session = cookie.Value
			break
		}
	}

	if delugeClient.Session == "" {
		return fmt.Errorf("session_id cookie not found after login")
	}

	return nil
}

func (delugeClient *DelugeClient) GetFinishedTorrents() ([]Torrent, error) {
	result, err := delugeClient.request(
		"core.get_torrents_status",
		[]any{
			map[string]any{
				"is_finished": []bool{true},
			},
			[]string{"hash", "total_size", "name", "time_since_download"},
		},
		true,
	)
	if err != nil {
		return nil, err
	}

	var torrentsMap map[string]TorrentDeluge
	if err := json.Unmarshal(result, &torrentsMap); err != nil {
		return nil, err
	}

	torrentsList := make([]Torrent, 0, len(torrentsMap))
	for _, torrent := range torrentsMap {
		torrentsList = append(torrentsList, torrent)
	}

	return torrentsList, nil
}

func (delugeClient *DelugeClient) RemoveTorrentsWithData(torrents []Torrent, dryRun bool) error {
	if dryRun {
		log.Log().Msg("remove torrents with dry-run")
		return nil
	}
	hashes := make([]string, 0, len(torrents))
	for _, torrent := range torrents {
		hashes = append(hashes, torrent.Id())
	}
	_, err := delugeClient.request("core.remove_torrents", []interface{}{hashes, true}, true)
	return err
}
