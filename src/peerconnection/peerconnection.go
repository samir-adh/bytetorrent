package peerconnection

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/samir-adh/bytetorrent/src/message"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	"github.com/samir-adh/bytetorrent/src/torrentfile"
	"github.com/samir-adh/bytetorrent/src/tracker"
	"github.com/ztrue/tracerr"
)

type PeerConnection struct {
	SelfId          [20]byte
	Peer            tracker.Peer
	TorrentFile     torrentfile.TorrentFile
	AvailablePieces []int
	//NetConn         net.Conn
}

func New(selfId [20]byte, peer tracker.Peer, torrentFile torrentfile.TorrentFile, netConn *net.Conn) (*PeerConnection, error) {

	connection := PeerConnection{
		SelfId:          selfId,
		Peer:            peer,
		TorrentFile:     torrentFile,
		AvailablePieces: nil,
		// NetConn:         netConn,
	}

	err := connection.handshakeExchange(netConn)
	if err != nil {
		return nil, err
	}

	// Receive bitfieldMessage
	bitfieldMessage, err := connection.receiveBitfield(netConn)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	fmt.Println("Received bitfield from peer:", bitfieldMessage.String())
	connection.AvailablePieces = getAvailablePieces(bitfieldMessage.Payload)

	// Send interested message
	if err := connection.sendInterested(netConn); err != nil {
		return nil, tracerr.Wrap(err)
	}

	// Receive unchoke message
	if err := connection.receiveUnchoke(netConn); err != nil {
		return nil, tracerr.Wrap(err)
	}
	fmt.Println("Received unchoke message from peer:", connection.Peer.String())

	return &connection, nil
}

func (connection *PeerConnection) handshakeExchange(netConn *net.Conn) error {
	// Send handshake
	sentHandshake, err := connection.SendHandShake(netConn)
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Receive handshake
	receivedHandshake, err := connection.ReceiveHandShake(netConn)
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Check the response
	if err := VerifyHandshake(sentHandshake, receivedHandshake); err != nil {
		return tracerr.Wrap(err)
	}
	fmt.Printf("Handshake verified with peer %s\n", connection.Peer.String())
	return nil
}

func getAvailablePieces(bitfield []byte) []int {
	list := make([]int, 0, 8*len(bitfield))
	for i, b := range bitfield {
		for pos := range 8 {
			if b&(1<<(7-pos)) != 0 {
				list = append(list, i*8+pos)
			}
		}
	}
	return list
}

func (p *PeerConnection) sendInterested(netConn *net.Conn) error {
	msg := &message.Message{
		Id:      message.MsgInterested,
		Length:  1,
		Payload: []byte{},
	}
	_, err := (*netConn).Write(msg.Serialize())
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

func (p *PeerConnection) receiveUnchoke(netConn *net.Conn) error {
	msg, err := message.Read(*netConn)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if msg.Id != message.MsgUnchoke {
		return fmt.Errorf("expected unchoke message, got %s", msg.Id.String())
	}
	fmt.Println("Received unchoke message from peer:", p.Peer.String())
	return nil
}

func (p *PeerConnection) receiveBitfield(netConn *net.Conn) (*message.Message, error) {
	msg, err := message.Read(*netConn)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if msg.Id != message.MsgBitfield {
		return nil, fmt.Errorf("expected bitfield message, got %s", msg.Id.String())
	}
	return msg, nil
}

func (p *PeerConnection) SendHandShake(netConn *net.Conn) (*HandShake, error) {
	handshakeMessage := HandShake{
		"BitTorrent protocol",
		p.TorrentFile.InfoHash,
		p.SelfId,
	}
	_, err := (*netConn).Write(handshakeMessage.Serialize())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &handshakeMessage, nil
}

func (p *PeerConnection) ReceiveHandShake(netConn *net.Conn) (*HandShake, error) {
	response := make([]byte, 68)
	_, err := (*netConn).Read(response)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	response_handshake := UnserializeHandshake(response)
	if err := VerifyHandshake(&HandShake{
		"BitTorrent protocol",
		p.TorrentFile.InfoHash,
		p.SelfId,
	}, &response_handshake); err != nil {
		return nil, tracerr.Wrap(err)
	}
	fmt.Printf("Handshake verified with peer %s\n", p.Peer.String())
	return &response_handshake, nil
}

func (h *HandShake) Serialize() []byte {
	handshakeMessage := make([]byte, 68)
	handshakeMessage[0] = 19
	copy(handshakeMessage[1:], "BitTorrent protocol")
	copy(handshakeMessage[28:], h.InfoHash[:])
	copy(handshakeMessage[48:], h.PeerId[:])
	return handshakeMessage
}

func UnserializeHandshake(handshake_bytes []byte) HandShake {
	handshake := HandShake{
		string(handshake_bytes[1:20]),
		[20]byte(handshake_bytes[28:48]),
		[20]byte(handshake_bytes[48:]),
	}
	return handshake
}

type HandShake struct {
	Protocol string
	InfoHash [20]byte
	PeerId   [20]byte
}

func VerifyHandshake(sent *HandShake, received *HandShake) error {
	if sent.Protocol != received.Protocol {
		err := fmt.Errorf("protocol mismatch, expected %s got %s", sent.Protocol, received.Protocol)
		return err
	}
	if sent.InfoHash != received.InfoHash {
		err := fmt.Errorf("infohash mismatch, expected %s got %s", sent.InfoHash, received.InfoHash)
		return err
	}
	return nil
}

func (p *PeerConnection) CanHandle(pieceIndex int) bool {
	for _, index := range p.AvailablePieces {
		if index == pieceIndex {
			return true
		}
	}
	return false
}

// func (p *PeerConnection) Download(toDownload piece.Piece) {
// 	fmt.Printf("Peer %s downloading piece %d", p.Peer.String(), toDownload.Index)
// 	time.Sleep(1 * time.Millisecond)
// }

func (p *PeerConnection) Download(piece pc.Piece, netConn *net.Conn) (*pc.PieceResult, error) {
	bytesDownloaded := 0
	pieceBuffer := make([]byte, piece.Length)
	for bytesDownloaded < piece.Length {
		blockSize := 16384
		if bytesDownloaded+blockSize > piece.Length {
			blockSize = piece.Length - bytesDownloaded
		}
		p.SendBlockRequest(piece, bytesDownloaded, blockSize, netConn)
		blockData, err := p.ReceiveBlock(piece.Index, bytesDownloaded, blockSize, netConn)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		copy(pieceBuffer[bytesDownloaded:], blockData)
	}
	return &pc.PieceResult{
		Index:   piece.Index,
		Payload: pieceBuffer,
	}, nil
}

func (p *PeerConnection) SendBlockRequest(piece pc.Piece, bytesDownloaded int, blockSize int, netConn *net.Conn) error {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(piece.Index))
	binary.BigEndian.PutUint32(payload[0:4], uint32(bytesDownloaded))
	binary.BigEndian.PutUint32(payload[0:4], uint32(blockSize))
	msg := message.Message{
		Id:      message.MsgRequest,
		Length:  uint32(len(payload) + 1),
		Payload: payload,
	}
	_, err := (*netConn).Write(msg.Serialize())
	if err != nil {
		return err
	}
	return nil
}

func (p *PeerConnection) ReceiveBlock(requestIndex int, requestBegin int, blockSize int, netConn *net.Conn) ([]byte, error) {
	response, err := message.Read(*netConn)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if response.Id != message.MsgPiece {
		return nil, tracerr.Errorf("Expected message with Id=7, got Id=%d", response.Id)
	}
	payload := response.Payload

	if index := int(binary.BigEndian.Uint32(payload[0:4])); index != requestIndex {
		return nil, tracerr.Errorf("Expected block from piece %d, got from piece %d", requestIndex, index)
	}

	if begin := int(binary.BigEndian.Uint32(payload[4:8])); begin != requestBegin {
		return nil, tracerr.Errorf("Begin mismatch")
	}
	blockData := payload[8:]

	if len(blockData) != blockSize {
		return nil, tracerr.Errorf("Mismatch in expected block size")
	}

	return blockData, nil
}
