package tracker

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/samir-adh/bytetorrent/src/torrentfile"
)

func generateTorrentFile() *torrentfile.TorrentFile {
	var infoHash [20]byte
	copy(infoHash[:], "01234567890123456789")

	piecesHash := make([][20]byte, 1)
	copy(piecesHash[0][:], "01234567890123456789")

	var testTorrentFile = torrentfile.TorrentFile{
		Announce:    "http://anounce.com:0000",
		InfoHash:    infoHash,
		PiecesHash:  piecesHash,
		PieceLength: 262144,
		Length:      1,
		Name:        "dummy",
	}

	return &testTorrentFile

}

func TestBuildTrackerRequest(t *testing.T) {
	torf := generateTorrentFile()
	var peerId [20]byte
	copy(peerId[:], "AbcdeAbcdeAbcdeAbcde")
	port := 1234
	actualRequestUrl, err := BuildTrackerRequest(torf, peerId, port)
	if err != nil {
		t.Error("failed to build tracker request.")
	}
	expectedRequestUrl := "http://anounce.com:0000?" +
		"compact=1&downloaded=0&info_hash=01234567890123456789" +
		"&left=1&peer_id=AbcdeAbcdeAbcdeAbcde&port=1234&uploaded=0"
	assertEqual(t, actualRequestUrl, expectedRequestUrl)
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got: %s, expected: %s", got, want)
	}
}

type mockHttpClient struct {
	response *http.Response
	err      error
}

func (m *mockHttpClient) Get(url string) (*http.Response, error) {
	return m.response, m.err
}

func TestFindPeers(t *testing.T) {
	peer1 := Peer{
		IpAdress: [4]byte{192, 168, 1, 1},
		Port:     [2]byte{0x1A, 0xE1},
	}
	peer2 := Peer{
		IpAdress: [4]byte{10, 0, 0, 1},
		Port:     [2]byte{0x1A, 0xE2},
	}
	peers := []Peer{peer1, peer2}
	body := `d8:intervali900e5:peers12:`
	for _, peer := range peers {
		body += string(peer.IpAdress[:])
		body += string(peer.Port[:])
	}
	body += `e`

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	client := &mockHttpClient{response: resp}
	peers, err := FindPeers("http://test.com", client)
	if err != nil {
		t.Errorf("test failed with error %s", err)
	}
	for i, peer := range peers {
		assertEqual(t, string(peer.IpAdress[:]), string(peers[i].IpAdress[:]))
		assertEqual(t, string(peer.Port[:]), string(peers[i].Port[:]))
	}
}
