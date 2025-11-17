package peerconnection

import (
	"encoding/binary"
	"fmt"
	"net"
	"github.com/samir-adh/bytetorrent/src/log"
	"github.com/samir-adh/bytetorrent/src/message"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	"github.com/samir-adh/bytetorrent/src/tracker"
	"github.com/ztrue/tracerr"
)

type PeerConnection struct {
	SelfId          [20]byte
	Peer            tracker.Peer
	AvailablePieces []int
	InfoHash        [20]byte
	logger          *log.Logger
	netConn         *net.Conn
}

func New(selfId [20]byte, peer tracker.Peer, infoHash [20]byte, netConn *net.Conn, logger *log.Logger) (*PeerConnection, error) {

	connection := PeerConnection{
		SelfId:          selfId,
		Peer:            peer,
		AvailablePieces: nil,
		InfoHash:        infoHash,
		logger:          logger,
		netConn:         netConn,
	}

	err := connection.handshakeExchange()
	if err != nil {
		return nil, err
	}

	// Receive bitfieldMessage
	bitfieldMessage, err := connection.receiveBitfield()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	logger.Println(log.HighVerbose, "Received bitfield from peer:", bitfieldMessage.String())
	connection.AvailablePieces = getAvailablePieces(bitfieldMessage.Payload)

	// Send interested message
	if err := connection.sendInterested(); err != nil {
		return nil, tracerr.Wrap(err)
	}

	// Receive unchoke message
	if err := connection.receiveUnchoke(); err != nil {
		return nil, tracerr.Wrap(err)
	}

	return &connection, nil
}

func (connection *PeerConnection) handshakeExchange() error {
	// Send handshake
	sentHandshake, err := connection.SendHandShake()
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Receive handshake
	receivedHandshake, err := connection.ReceiveHandShake()
	if err != nil {
		return tracerr.Wrap(err)
	}

	// Check the response
	if err := VerifyHandshake(sentHandshake, receivedHandshake); err != nil {
		return tracerr.Wrap(err)
	}
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

func (p *PeerConnection) sendInterested() error {
	msg := &message.Message{
		Id:      message.MsgInterested,
		Length:  1,
		Payload: []byte{},
	}
	_, err := (*p.netConn).Write(msg.Serialize())
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

func (p *PeerConnection) receiveUnchoke() error {
	msg, err := message.Read(*p.netConn)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if msg.Id != message.MsgUnchoke {
		return fmt.Errorf("expected unchoke message, got %s", msg.Id.String())
	}
	p.logger.Println(log.HighVerbose, "Received unchoke message from peer:", p.Peer.String())
	return nil
}

func (p *PeerConnection) receiveBitfield() (*message.Message, error) {
	msg, err := message.Read(*p.netConn)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if msg.Id != message.MsgBitfield {
		return nil, fmt.Errorf("expected bitfield message, got %s", msg.Id.String())
	}
	return msg, nil
}

func (p *PeerConnection) SendHandShake() (*HandShake, error) {
	handshakeMessage := HandShake{
		"BitTorrent protocol",
		p.InfoHash,
		p.SelfId,
	}
	_, err := (*p.netConn).Write(handshakeMessage.Serialize())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &handshakeMessage, nil
}

func (p *PeerConnection) ReceiveHandShake() (*HandShake, error) {
	response := make([]byte, 68)
	_, err := (*p.netConn).Read(response)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	response_handshake := UnserializeHandshake(response)
	if err := VerifyHandshake(&HandShake{
		"BitTorrent protocol",
		p.InfoHash,
		p.SelfId,
	}, &response_handshake); err != nil {
		return nil, tracerr.Wrap(err)
	}
	p.logger.Printf(log.HighVerbose, "Handshake verified with peer %s\n", p.Peer.String())
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

type block struct {
	Index  int
	Offset int
	Data   []byte
}

func (p *PeerConnection) Download(piece *pc.Piece) (*pc.PieceResult, error) {
	defaultBlockSize := 16384
	blocksCount := piece.Length / defaultBlockSize
	if piece.Length%defaultBlockSize != 0 {
		blocksCount++
	}
	downloadedBlocks := 0
	for i := range blocksCount {
		offset := i * defaultBlockSize
		blockSize := defaultBlockSize
		if offset+defaultBlockSize > piece.Length {
			blockSize = piece.Length - offset
		}
		if err := p.sendBlockRequest(piece, offset, blockSize); err != nil {
			return nil, tracerr.Wrap(err)
		}
	}

	pieceBuffer := make([]byte, piece.Length)

	for downloadedBlocks < blocksCount {
		response, err := message.Read(*p.netConn)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if response.Id != message.MsgPiece {
			switch response.Id {
			default:
				return nil, fmt.Errorf("expected message id %d, got %d", message.MsgPiece, response.Id)
			}
		}

		block := parseBlockData(response.Payload)
		copy(pieceBuffer[block.Offset:], block.Data)
		downloadedBlocks++
	}

	return &pc.PieceResult{
		Index:   piece.Index,
		Payload: pieceBuffer,
		State:   pc.Downloaded,
	}, nil

}

func parseBlockData(payload []byte) *block {
	index := binary.BigEndian.Uint32(payload[0:4])
	offset := binary.BigEndian.Uint32(payload[4:8])
	data := make([]byte, len(payload[8:]))
	copy(data, payload[8:])
	return &block{
		Index:  int(index),
		Offset: int(offset),
		Data:   data,
	}

}

func (p *PeerConnection) sendBlockRequest(piece *pc.Piece, bytesDownloaded int, blockSize int) error {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(piece.Index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(bytesDownloaded))
	binary.BigEndian.PutUint32(payload[8:12], uint32(blockSize))
	msg := message.Message{
		Id:      message.MsgRequest,
		Length:  uint32(len(payload) + 1),
		Payload: payload,
	}
	_, err := (*p.netConn).Write(msg.Serialize())
	if err != nil {
		return err
	}
	return nil
}
