package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type ID uint8

// the keep-alive message is a message with zero bytes, specified with the length prefix set to zero
// keep-alive: <len=0000>
const (
	MsgChoke         ID = 0 // choke: <len=0001><id=0>
	MsgUnchoke       ID = 1 // unchoke: <len=0001><id=1>
	MsgInterested    ID = 2 // interested: <len=0001><id=2>
	MsgNotInterested ID = 3 // not interested: <len=0001><id=3>
	MsgHave          ID = 4 // have: <len=0005><id=4><piece index>
	MsgBitfield      ID = 5 // bitfield: <len=0001+X><id=5><bitfield>
	MsgRequest       ID = 6 // request: <len=0013><id=6><index><begin><length>
	MsgPiece         ID = 7 // piece: <len=0009+X><id=7><index><begin><block>
	MsgCancel        ID = 8 // cancel: <len=0013><id=8><index><begin><length>
	MsgPort          ID = 9 // port: <len=0003><id=9><listen-port>
)

// Message stores ID and payload of a message
type Message struct { // message: <length prefix><message ID><payload>
	Prefix    uint32
	MessageID ID
	Payload   []byte
}

// FormatRequest creates a `request` message
func FormatRequest(index int, begin int, length int) *Message {
	var payload = make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{
		MessageID: MsgRequest,
		Payload:   payload,
		Prefix:    uint32(1 + 12),
	}
}

// FormatHave creates a `have` message
func FormatHave(index int) *Message {
	var payload = make([]byte, 4)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	return &Message{
		Prefix:    uint32(1 + 4),
		MessageID: MsgHave,
		Payload:   payload,
	}
}

// Serialize serializes a message into a buffer of the form <length prefix><message ID><payload>
// interprets `nil` as a keep-alive message
func (m *Message) Serialize() []byte {
	if m.Prefix == uint32(0) {
		return make([]byte, 4)
	}
	m.Prefix = uint32(len(m.Payload) + 1) // +1 for <message ID>
	binary.BigEndian.PutUint32(m.Payload[0:4], m.Prefix)
	var buf = make([]byte, 4+m.Prefix)
	binary.BigEndian.PutUint32(buf[0:4], m.Prefix)
	buf[4] = byte(m.MessageID)
	_ = copy(buf[5:], m.Payload)
	return buf
}

// ParseHave parses a `have` message
func ParseHave(msg *Message) (int, error) {
	if msg.MessageID != MsgHave {
		return 0, fmt.Errorf("expected have (<id=%d>) message, got %d", MsgHave, msg.MessageID)
	}
	if len(msg.Payload) != 5-1 {
		return 0, fmt.Errorf("expected <piece index> length %d, got %d", 5-1, len(msg.Payload))
	}
	return int(binary.BigEndian.Uint32(msg.Payload[0:4])), nil
}

// ParsePiece parses a `piece` message and copies its payload into a buffer
func ParsePiece(index int, msg *Message, buf []byte) (int, error) {
	if msg.MessageID != MsgPiece {
		return 0, fmt.Errorf("expected piece (<id=%d>) message, got <id=%d>", MsgPiece, msg.MessageID)
	}
	if len(msg.Payload) < 9-1 {
		return 0, fmt.Errorf("payload too short, %d < %d", len(msg.Payload), 9-1)
	}
	var parsedIndex = int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if index != parsedIndex {
		return 0, fmt.Errorf("expected <index>=%d, got %d", index, parsedIndex)
	}
	var begin = int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("offset <begin> too high, %d >= %d", begin, len(buf))
	}
	var data = msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data <block> %d too long for offset <begin> %d with length %d", len(data), begin, len(buf))
	}
	return copy(buf[begin:], data), nil
}

// Read parses a message from a stream
func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == uint32(0) { // keep-alive message
		return &Message{Prefix: length}, nil
	}
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}
	return &Message{
		Prefix:    length,
		MessageID: ID(messageBuf[0]),
		Payload:   messageBuf[1:],
	}, nil
}

func (m *Message) name() string {
	if m == nil {
		return "malformed message"
	}
	if m.Prefix == uint32(0) {
		return "keep-alive"
	}
	switch m.MessageID {
	case MsgChoke:
		return "choke"
	case MsgUnchoke:
		return "unchoke"
	case MsgInterested:
		return "interested"
	case MsgNotInterested:
		return "not interested"
	case MsgHave:
		return "have"
	case MsgBitfield:
		return "bitfield"
	case MsgRequest:
		return "request"
	case MsgPiece:
		return "piece"
	case MsgCancel:
		return "cancel"
	case MsgPort:
		return "port"
	default:
		return fmt.Sprintf("unknown (<id=%d>)", m.MessageID)
	}
}

func (m *Message) String() string {
	if m == nil {
		return m.name()
	} else {
		return fmt.Sprintf("<message ID> %d (%s), <payload> length %d", m.MessageID, m.name(), len(m.Payload))
	}
}
