package tracker

import (
	"crypto/rand"
	"fmt"
	"github.com/jackpal/bencode-go"
	tf "github.com/samir-adh/bytetorrent/torrentfile"
	"github.com/ztrue/tracerr"
	"net/http"
	"net/url"
)

type Peer struct {
	IpAdress [4]byte ""
	Port     [2]byte
}

type TrackerResponse struct {
	Interval   int
	Complete   int
	Incomplete int
	Peers      []Peer
}

type BencondeTrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func RandomPeerId() ([20]byte, error) {
	var buffer [20]byte
	bytes_written, err := rand.Read(buffer[:])
	if bytes_written != 20 {
		return buffer, fmt.Errorf("wrong amount of bytes written, expected: 20 but go ")
	}
	if err != nil {
		return buffer, tracerr.Wrap(err)
	}
	return buffer, nil
}

func UrlEncodedInfoHash(tor *tf.TorrentFile) (string, error) {
	return string(tor.InfoHash[:]), nil
}

// Builds the url to send to the tracker in order to receive peers,
// the 'peerId' argument corresponds to the the id of our client
func BuildTrackerRequest(tor *tf.TorrentFile, peerId [20]byte, port int) (string, error) {
	infoHash, err := UrlEncodedInfoHash(tor) // encode the infohash in UTF-8
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	params := url.Values{
		"info_hash":  []string{infoHash}, // URL-encode needed!
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{fmt.Sprintf("%d", tor.Length)},
		"compact":    []string{"1"}, // request compact peer list
	}
	urlStr := tor.Announce + "?" + params.Encode()
	return urlStr, nil
}

func ParsePeers(peers []byte) ([]Peer, error) {
	peerCount := len(peers) / 6
	peerList := make([]Peer, peerCount)
	for i := range peerCount {
		peerList[i].IpAdress = [4]byte{peers[i*6], peers[i*6+1], peers[i*6+2], peers[i*6+3]}
		peerList[i].Port = [2]byte{peers[i*6+4], peers[i*6+5]}
	}
	return peerList, nil
}

type HttpClient interface {
	Get(url string) (*http.Response, error)
}

func FindPeers(fullURL string, client HttpClient) ([]Peer, error) {
	resp, err := client.Get(fullURL)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer resp.Body.Close()
	var tr_response BencondeTrackerResponse
	err = bencode.Unmarshal(resp.Body, &tr_response)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	peers, err := ParsePeers([]byte(tr_response.Peers))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return peers, nil
}

func (peer *Peer) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", peer.IpAdress[0], peer.IpAdress[1], peer.IpAdress[2], peer.IpAdress[3], (int(peer.Port[0])<<8)+int(peer.Port[1]))
}
