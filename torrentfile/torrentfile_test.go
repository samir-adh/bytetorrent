package torrentfile

import (
	"testing"
)

func TestOpenTorrentFile(t *testing.T) {
	filepath := "testdata/cosmos-laundromat.torrent"
	tf, err := OpenTorrentFile(filepath)
	if err != nil {
		t.Errorf("%v",err)
		return
	}
	got_announce := tf.Announce
	expected_announce := "udp://tracker.leechers-paradise.org:6969"
	assertEqual(t, got_announce, expected_announce)
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got: %s, expected: %s", got, want)
	}
}
