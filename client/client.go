package client

import (
	"io"
	"net"
)

// Client is a connection with a peer
type Client struct { // TODO: currently only TCP is supported
	Conn net.Conn
}

// Handshake represents a special message that a peer uses to identify itself
type Handshake struct {
	PstrLength byte
	Pstr       string
	Reserved   [8]byte
	InfoHash   [20]byte
	PeerID     [20]byte
}

// Serialize serializes the handshake to a buffer
func (h *Handshake) Serialize() ([]byte, error) {
	length := 1 + len(h.Pstr) + 8 + 20 + 20
	buf := make([]byte, 1+len(h.Pstr)+8+20+20)
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
