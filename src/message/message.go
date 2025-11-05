package message

import (
	"encoding/binary"
	"io"
	"strconv"

	"github.com/ztrue/tracerr"
)

type Message struct {
	Id      messageId
	Length  uint32
	Payload []byte
}

type messageId uint8

const (
	// MsgChoke chokes the receiver
	MsgChoke messageId = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke messageId = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested messageId = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested messageId = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave messageId = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield messageId = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest messageId = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece messageId = 7
	// MsgCancel cancels a request
	MsgCancel messageId = 8
)

func Read(r io.Reader) (*Message, error) {
	// Read message length
	buf_length := make([]byte, 4)
	_, err := io.ReadFull(r, buf_length)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	length := binary.BigEndian.Uint32(buf_length)
	if length <= 0 {
		return nil, tracerr.Errorf("Failed to read message with length %d", length)
	}
	// Read the rest of the message
	buf_message := make([]byte, length)
	_, err = io.ReadFull(r, buf_message) // the first 4 bytes were already read 
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	message_id := uint8(buf_message[0])
	payload := buf_message[1:]
	return &Message{
		Id:      messageId(message_id),
		Length:  uint32(length),
		Payload: payload,
	}, nil
}

// String returns a string representation of the messageId
func (m *messageId) String() string {
	var message_type string
	switch *m {
	case MsgChoke:
		message_type = "Choke"
	case MsgUnchoke:
		message_type = "Unchoke"
	case MsgInterested:
		message_type = "Interested"
	case MsgNotInterested:
		message_type = "Not Interested"
	case MsgHave:
		message_type = "Have"
	case MsgBitfield:
		message_type = "Bitfield"
	case MsgRequest:
		message_type = "Request"
	case MsgPiece:
		message_type = "Piece"
	case MsgCancel:
		message_type = "Cancel"
	default:
		message_type = "Unknown"
	}
	return message_type
}

func (m *Message) String() string {
	return "Message{" +
		"Id: " + m.Id.String() +
		", Length: " + strconv.Itoa(int(m.Length)) +
		//", Payload: " + string(m.Payload) +
		"}"
}

func (m *Message) Serialize() []byte {
	buf := make([]byte, 4+m.Length)
	binary.BigEndian.PutUint32(buf[0:4], m.Length)
	buf[4] = byte(m.Id)
	copy(buf[5:], m.Payload)
	return buf
}
