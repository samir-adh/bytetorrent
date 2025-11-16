package torrentfile

import (
	"fmt"
	"testing"
)

var testBencodeTorrent = BencodeTorrent{
	Announce:     "anounce",
	AnnounceList: nil,
	Info: BencodeInfo{
		Name:        "test-file.txt",
		Length:      1024,
		PieceLength: 32768,
		Pieces:      "01234567890123456789", // 20-byte dummy SHA1 hash
	},
}

func TestOpenTorrentFile(t *testing.T) {
	filepath := "testdata/cosmos-laundromat.torrent"
	tf, err := OpenTorrentFile(filepath)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	got_announce := tf.Announce
	expected_announce := "udp://tracker.leechers-paradise.org:6969"
	assertEqual(t, got_announce, expected_announce)
}

func TestInfoHash(t *testing.T) {
	bto := testBencodeTorrent
	expectedHashHex := "bc72e87aba71343c61ffe8469ed32bbfd70f8eb0"
	actualHash, err := bto.InfoHash()
	if err != nil {
		t.Errorf("Failed calculating the info hash")
	}
	actualHashHex := fmt.Sprintf("%x", actualHash)

	assertEqual(t, actualHashHex, expectedHashHex)
}

func TestPiecesHash(t *testing.T) {
	bto := testBencodeTorrent
	expectedHashList := []string{"01234567890123456789"}
	actualHashByteList, err := bto.PiecesHash()
	if err != nil {
		t.Errorf("Failed recovering the pieces hashs")
	}
	for i, actualHash := range actualHashByteList {
		actualHashHex := fmt.Sprintf("%s", actualHash)
		expectedHash := expectedHashList[i]
		assertEqual(t, actualHashHex, expectedHash)
	}

}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got: %s, expected: %s", got, want)
	}
}
