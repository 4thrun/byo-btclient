package peers

import (
	"fmt"
	"io"
)

// Handshake represents a special message that a peer uses to identify itself
type Handshake struct { // handshake: <pstrlen><pstr><reserved><info_hash><peer_id>
	PstrLength byte
	Pstr       string
	Reserved   [8]byte
	InfoHash   [20]byte
	PeerID     [20]byte
}

// Serialize serializes the handshake to a buffer
func (h *Handshake) Serialize() ([]byte, error) {
	length := 1 + len(h.Pstr) + 8 + 20 + 20
	buf := make([]byte, length)
	buf[0] = h.PstrLength
	copied := 1
	copied += copy(buf[copied:], h.Pstr)
	copied += copy(buf[copied:], h.Reserved[:])
	copied += copy(buf[copied:], h.InfoHash[:])
	copied += copy(buf[copied:], h.PeerID[:])
	if copied != length {
		return nil, io.ErrShortBuffer
	}
	return buf, nil
}

// New creates a new handshake with the standard pstr
func New(infoHash [20]byte, peerID [20]byte, reserved [8]byte) (*Handshake, error) {
	var standardPstr = "BitTorrent protocol"
	var length = len(standardPstr)
	if length > int(^byte(0)) {
		return nil, io.ErrShortBuffer
	}
	return &Handshake{
		Pstr:       standardPstr,
		PstrLength: byte(length),
		Reserved:   reserved,
		InfoHash:   infoHash,
		PeerID:     peerID,
	}, nil
}

// Read parses a handshake from a stream
func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrLen := int(lengthBuf[0])
	if pstrLen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}
	handshakeBuf := make([]byte, pstrLen+20+20+8)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}
	var infoHash [20]byte
	var peerID [20]byte
	_ = copy(infoHash[:], handshakeBuf[pstrLen+8:pstrLen+8+20])
	_ = copy(peerID[:], handshakeBuf[pstrLen+8+20:])
	h := &Handshake{
		PstrLength: lengthBuf[0],
		Pstr:       string(handshakeBuf[0:pstrLen]),
		InfoHash:   infoHash,
		PeerID:     peerID,
	}
	return h, nil
}
