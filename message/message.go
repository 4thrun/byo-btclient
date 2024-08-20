package peers

import (
	"encoding/binary"
	"io"
)

type MessageID uint8

// the keep-alive message is a message with zero bytes, specified with the length prefix set to zero
// keep-alive: <len=0000>
const (
	MsgChoke         MessageID = 0 // choke: <len=0001><id=0>
	MsgUnchoke       MessageID = 1 // unchoke: <len=0001><id=1>
	MsgInterested    MessageID = 2 // interested: <len=0001><id=2>
	MsgNotInterested MessageID = 3 // not interested: <len=0001><id=3>
	MsgHave          MessageID = 4 // have: <len=0005><id=4><piece index>
	MsgBitfield      MessageID = 5 // bitfield: <len=0001+X><id=5><bitfield>
	MsgRequest       MessageID = 6 // request: <len=0013><id=6><index><begin><length>
	MsgPiece         MessageID = 7 // piece: <len=0009+X><id=7><index><begin><block>
	MsgCancel        MessageID = 8 // cancel: <len=0013><id=8><index><begin><length>
	MsgPort          MessageID = 9 // port: <len=0003><id=9><listen-port>
)

// Message stores ID and payload of a message
type Message struct { // message: <length prefix><message ID><payload>
	Prefix    [4]byte
	MessageID MessageID
	Payload   []byte
}

// Serialize serializes a message into a buffer of the form <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	payloadLen := uint32(len(m.Payload))
	var buf = make([]byte, 4+1+payloadLen)
	binary.BigEndian.PutUint32(buf[0:4], payloadLen)
	buf[4] = byte(m.MessageID)
	_ = copy(buf[5:], m.Payload)
	return buf
}

// Read parses a message from a stream
// Returns `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {

}
