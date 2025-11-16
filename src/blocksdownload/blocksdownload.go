package blocksdownload

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/samir-adh/bytetorrent/src/message"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	"github.com/ztrue/tracerr"
)

type block struct {
	Index  int
	Offset int
	Data   []byte
}


func DownloadPiece(piece *pc.Piece, netConn *net.Conn) (*pc.PieceResult, error) {
	defaultBlockSize := 16384
	blocksCount := (piece.Length / defaultBlockSize)
	for i := range blocksCount {
		offset := i * defaultBlockSize
		blockSize := 16384
		if offset+defaultBlockSize > piece.Length {
			blockSize = piece.Length - offset
		}
		if err := sendBlockRequest(piece, offset, blockSize, netConn); err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	
	pieceBuffer := make([]byte, piece.Length)
	
	for range blocksCount {
		// fmt.Printf("%dth block from piece %d\n", i+1, piece.Index)
		response, err := message.Read(*netConn)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if response.Id != message.MsgPiece {
			return nil, fmt.Errorf("expected message id %d, got %d", message.MsgPiece, response.Id)
		}

		block := parseBlockData(response.Payload)
		copy(pieceBuffer[block.Offset:], block.Data)
	}

	return &pc.PieceResult{
		Index:   piece.Index,
		Payload: pieceBuffer,
		Error:   nil,
	}, nil

}

func parseBlockData(payload []byte) *block {
	index := binary.BigEndian.Uint32(payload[0:4])
	offset := binary.BigEndian.Uint32(payload[4:8])
	data := payload[8:]
	return &block{
		Index:  int(index),
		Offset: int(offset),
		Data:   data,
	}

}

func sendBlockRequest(piece *pc.Piece, bytesDownloaded int, blockSize int, netConn *net.Conn) error {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(piece.Index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(bytesDownloaded))
	binary.BigEndian.PutUint32(payload[8:12], uint32(blockSize))
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