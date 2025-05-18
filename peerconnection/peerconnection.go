package peerconnection

import (
	"fmt"
	"net"
	"time"

	tf "github.com/samir-adh/bytetorrent/torrentfile"
	tr "github.com/samir-adh/bytetorrent/tracker"
	"github.com/ztrue/tracerr"
)

type PeerConnection struct {
	SelfId      [20]byte
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
		return tracerr.Wrap(err)
	}
	h := HandShake{
		"BitTorrent protocol",
		connection.TorrentFile.InfoHash,
		connection.SelfId,
	}
	// handshakeMessage := BuildHandshakeMessage(connection.SelfId, string(connection.TorrentFile.InfoHash[:]))
	_, err = conn.Write(h.Serialize())
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Read the response
	response := make([]byte, 68)
	_, err = conn.Read(response)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Check the response
	fmt.Printf("Received response: %x\n", response)
	return nil
}

func (h *HandShake) Serialize() []byte {
	handshakeMessage := make([]byte, 68)
	handshakeMessage[0] = 19
	copy(handshakeMessage[1:], "BitTorrent protocol")
	copy(handshakeMessage[28:], h.InfoHash[:])
	copy(handshakeMessage[48:], h.PeerId[:])
	return handshakeMessage
}

type HandShake struct {
	Protocol string
	InfoHash [20]byte
	PeerId   [20]byte
}
