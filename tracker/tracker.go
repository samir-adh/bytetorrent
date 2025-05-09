package tracker

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"

	tf "github.com/samir-adh/bytetorrent/torrentfile"
)

type Peer struct {
	Id       string
	IpAdress [4]byte ""
	Port     [2]byte
}

type TrackerResponse struct {
	Interval   int
	Complete   int
	Incomplete int
	Peers      []Peer
}

func GeneratePeerId() (string, error) {
	buffer := make([]byte, 20)
	bytes_written, err := rand.Read(buffer)
	if bytes_written != 20 {
		return "", fmt.Errorf("wrong amount of bytes written, expected: 20 but go ")
	}
	if err != nil {
		return "", fmt.Errorf("failed to write peer id buffer: %v", err)
	}
	return string(buffer), nil
}

func UrlEncodedInfoHash(tor *tf.TorrentFile) (string, error) {
	return string(tor.InfoHash[:]), nil
}

// func BuildTrackerRequest(baseUrl string, infoHash [20]byte, peerId string, port int, length int) string {
func BuildTrackerRequest(tor *tf.TorrentFile, peerId string, port int) (string, error) {
	infoHash, err := UrlEncodedInfoHash(tor)
	if err != nil {
		return "", fmt.Errorf("failed to encode infohash in url format: %v", err)
	}
	params := url.Values{
		"info_hash":  []string{infoHash}, // URL-encode needed!
		"peer_id":    []string{peerId},
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{fmt.Sprintf("%d", tor.Length)},
		"compact":    []string{"1"}, // request compact peer list
	}
	urlStr := tor.Announce + "?" + params.Encode()
	return urlStr, nil
}
func ConnectToTracker(fullURL string) (string, error) {
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("error sending request to tracker: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Printf("Response code : %d\n", resp.StatusCode)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %v", err)
	}
	return string(body), nil
}
