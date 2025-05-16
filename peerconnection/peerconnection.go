package peerconnection

import (
	"fmt"
	"net"
	"time"

	tf "github.com/samir-adh/bytetorrent/torrentfile"
	tr "github.com/samir-adh/bytetorrent/tracker"
)

type PeerConnection struct {
	SelfId      string
	Peer        tr.Peer
	TorrentFile *tf.TorrentFile
}

func ConnectToPeer(connection *PeerConnection) error {
	conn, err := net.DialTimeout(
		"tcp",
		connection.Peer.String(),
		5*time.Second,
	)
	if err != nil {
		return err
	}
	handshakeMessage := BuildHandshakeMessage(connection.SelfId, string(connection.TorrentFile.InfoHash[:]))
	_, err = conn.Write(handshakeMessage)
	if err != nil {
		return err
	}
	// Read the response
	response := make([]byte, 68)
	_, err = conn.Read(response)
	if err != nil {
		return err
	}
	// Check the response
	fmt.Printf("Received response: %x\n", response)
	return nil
}

func BuildHandshakeMessage(peerId string, infoHash string) []byte {
	var handshakeMessage = make([]byte, 68)
	handshakeMessage[0] = 19
	copy(handshakeMessage[1:], "BitTorrent protocol")
	copy(handshakeMessage[28:], []byte(infoHash))
	copy(handshakeMessage[48:], []byte(peerId))
	return handshakeMessage
}
