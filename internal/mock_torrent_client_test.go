package internal

type MockTorrentClient struct {
	CheckConnectionCalled        bool
	CheckConnectionError         error
	LoginCalled                  bool
	LoginError                   error
	GetFinishedTorrentsCalled    bool
	GetFinishedTorrentsResult    []Torrent
	GetFinishedTorrentsError     error
	RemoveTorrentsWithDataCalled bool
	RemoveTorrentsWithDataInput  []Torrent
	RemoveTorrentsWithDataDryRun bool
	RemoveTorrentsWithDataError  error
}

func NewMockTorrentClient() *MockTorrentClient {
	return &MockTorrentClient{
		GetFinishedTorrentsResult: []Torrent{},
	}
}

func (m *MockTorrentClient) CheckConnection() {
	m.CheckConnectionCalled = true
}

func (m *MockTorrentClient) Login() error {
	m.LoginCalled = true
	return m.LoginError
}

func (m *MockTorrentClient) GetFinishedTorrents() ([]Torrent, error) {
	m.GetFinishedTorrentsCalled = true
	return m.GetFinishedTorrentsResult, m.GetFinishedTorrentsError
}

func (m *MockTorrentClient) RemoveTorrentsWithData(torrents []Torrent, dryRun bool) error {
	m.RemoveTorrentsWithDataCalled = true
	m.RemoveTorrentsWithDataInput = torrents
	m.RemoveTorrentsWithDataDryRun = dryRun
	return m.RemoveTorrentsWithDataError
}
