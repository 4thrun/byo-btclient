package client

import (
	"bittorrent-client-go/bitfield"
	"bittorrent-client-go/handshake"
	"bittorrent-client-go/message"
	"bittorrent-client-go/peers"
	"bytes"
	"fmt"
	"net"
	"time"
)

// Unless specified otherwise, all integers in the peer wire protocol are encoded as four byte big-endian values.
// This includes the length prefix on all messages that come after the handshake.

// clientInfo maintains state information for each connection that it has with a remote peer
type clientInfo struct {
	AmChoking      bool
	AmInterested   bool
	PeerChoking    bool
	PeerInterested bool
}

// Client is a connection with a peer
type Client struct {
	Conn       net.Conn
	ClientInfo clientInfo
	Bitfield   bitfield.Bitfield
	Peer       peers.Peer
	InfoHash   [20]byte
	PeerID     [20]byte
}

// New connects with a peer, completes a handshake, and receives a handshake
func New(peer peers.Peer, peerID [20]byte, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), time.Second*5) // TODO: currently only TCP is supported
	if err != nil {
		return nil, err
	}
	_, err = completeHandShake(conn, infoHash, peerID)
	if err != nil {
		return nil, err
	}
	bf, err := recvBitfield(conn)
	if err != nil { // ask another peer later
		_ = conn.Close()
		return nil, err
	}
	return &Client{
		Conn: conn,
		ClientInfo: clientInfo{
			AmChoking:      true,
			AmInterested:   false,
			PeerChoking:    true,
			PeerInterested: false,
		},
		Bitfield: bf,
		Peer:     peer,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}

func completeHandShake(conn net.Conn, infoHash [20]byte, peerID [20]byte) (*handshake.Handshake, error) {
	_ = conn.SetDeadline(time.Now().Add(time.Second * 5))
	defer func(conn net.Conn, t time.Time) {
		_ = conn.SetDeadline(t)
	}(conn, time.Time{}) // disable the deadline
	req, err := handshake.New(infoHash, peerID, [8]byte{0, 0, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return nil, err
	}
	serialized, err := req.Serialize()
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(serialized)
	if err != nil {
		return nil, err
	}
	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("expected <info_hash> %x, got %x", infoHash, res.InfoHash)
	}
	return res, nil
}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	_ = conn.SetDeadline(time.Now().Add(time.Second * 5))
	defer func(conn net.Conn, t time.Time) {
		_ = conn.SetDeadline(t)
	}(conn, time.Time{}) // disable the deadline
	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("expected message, got %v", msg)
		return nil, err
	}
	if msg.MessageID != message.MsgBitfield {
		err := fmt.Errorf("expected bitfield (<id=5>), got <message ID> %v", msg.MessageID)
		return nil, err
	}
	return msg.Payload, nil
}

// Read reads and consumes a message from the connection
func (c *Client) Read() (msg *message.Message, err error) {
	msg, err = message.Read(c.Conn)
	return msg, err
}

// SendRequest sends a `request` message to a peer
func (c *Client) SendRequest(index, begin, length int) error {
	var req = message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

// SendInterested sends an `interested` message to a peer
func (c *Client) SendInterested() error {
	var msg = message.Message{MessageID: message.MsgInterested, Prefix: uint32(1)}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotInterested sends a `not interested` message to a peer
func (c *Client) SendNotInterested() error {
	var msg = message.Message{MessageID: message.MsgNotInterested, Prefix: uint32(1)}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke sends an `unchoke` message to a peer
func (c *Client) SendUnchoke() error {
	var msg = message.Message{MessageID: message.MsgUnchoke, Prefix: uint32(1)}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave sends a `have` message to a peer
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
