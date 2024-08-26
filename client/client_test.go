package client

import (
	"bittorrent-client-go/handshake"
	"bittorrent-client-go/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

// createClientAndServer: helper function
func createClientAndServer(t *testing.T) (clientConn net.Conn, serverConn net.Conn) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)
	flag := make(chan struct{}) // net.Dial doesn't block, so we need this signalling channel to make sure we don't return before `serverConn` is ready
	go func() {
		defer close(flag)
		serverConn, err = ln.Accept()
		require.Nil(t, err)
		flag <- struct{}{}
	}()
	clientConn, err = net.Dial("tcp", ln.Addr().String())
	<-flag
	return
}

func TestCompleteHandshake(t *testing.T) {
	var tests = map[string]struct {
		clientInfoHash  [20]byte
		clientPeerID    [20]byte
		serverHandshake []byte
		output          *handshake.Handshake
		failure         bool
	}{
		"successful handshake": {
			clientInfoHash:  [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
			clientPeerID:    [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			serverHandshake: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			output: &handshake.Handshake{
				Pstr:       "BitTorrent protocol",
				PstrLength: byte(19),
				Reserved:   [8]byte{0, 0, 0, 0, 0, 0, 0, 0},
				InfoHash:   [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
				PeerID:     [20]byte{45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			},
			failure: false,
		},
		"wrong <info_hash>": {
			clientInfoHash:  [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
			clientPeerID:    [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			serverHandshake: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 0xde, 0xe8, 0x6a, 0x7f, 0xa6, 0xf2, 0x86, 0xa9, 0xd7, 0x4c, 0x36, 0x20, 0x14, 0x61, 0x6a, 0x0f, 0xf5, 0xe4, 0x84, 0x3d, 45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			output:          nil,
			failure:         true,
		},
	}
	for _, test := range tests {
		clientConn, serverConn := createClientAndServer(t)
		_, _ = serverConn.Write(test.serverHandshake)
		h, err := completeHandShake(clientConn, test.clientInfoHash, test.clientPeerID)
		if test.failure {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, h, test.output)
		}
	}
}

func TestClient_Read(t *testing.T) {
	clientConn, serverConn := createClientAndServer(t)
	client := Client{Conn: clientConn}
	msgBytes := []byte{
		0x00, 0x00, 0x00, 0x05,
		4,
		0x00, 0x00, 0x05, 0x3c,
	}
	expected := &message.Message{
		MessageID: message.MsgHave,
		Payload:   []byte{0x00, 0x00, 0x05, 0x3c},
		Prefix:    uint32(1 + 4),
	}
	_, err := serverConn.Write(msgBytes)
	require.Nil(t, err)
	msg, _ := client.Read()
	assert.Equal(t, msg, expected)
}

func TestClient_SendRequest(t *testing.T) {
	var clientConn, serverConn net.Conn = createClientAndServer(t)
	var client = Client{Conn: clientConn}
	var err = client.SendRequest(1, 2, 3)
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x0d,
		6,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03,
	}
	var buf = make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}

func TestClient_SendInterested(t *testing.T) {
	var clientConn, serverConn net.Conn = createClientAndServer(t)
	client := Client{Conn: clientConn}
	err := client.SendInterested()
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x01,
		2,
	}
	var buf = make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}

func TestClient_SendNotInterested(t *testing.T) {
	var clientConn, serverConn net.Conn = createClientAndServer(t)
	var client = Client{Conn: clientConn}
	var err = client.SendNotInterested()
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x01,
		3,
	}
	buf := make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, expected, buf)
}

func TestClient_SendUnchoke(t *testing.T) {
	var clientConn, serverConn net.Conn = createClientAndServer(t)
	var client = Client{Conn: clientConn}
	err := client.SendUnchoke()
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x01,
		1,
	}
	buf := make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, buf, expected)
}

func TestClient_SendHave(t *testing.T) {
	var clientConn, serverConn net.Conn = createClientAndServer(t)
	var client = Client{Conn: clientConn}
	var err = client.SendHave(1340)
	assert.Nil(t, err)
	expected := []byte{
		0x00, 0x00, 0x00, 0x05,
		4,
		0x00, 0x00, 0x05, 0x3c,
	}
	var buf = make([]byte, len(expected))
	_, err = serverConn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, buf, expected)
}
