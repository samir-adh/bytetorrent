package tracker

import (
	"github.com/samir-adh/bytetorrent/torrentfile"
	"testing"
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
