package peers

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestPeer_String(t *testing.T) {
	var tests = []struct {
		input  Peer
		output string
	}{
		{
			input:  Peer{IP: net.IP{127, 0, 0, 1}, Port: 8080},
			output: "127.0.0.1:8080",
		},
		{
			input:  Peer{IP: net.IP{192, 168, 0, 1}, Port: 1024},
			output: "192.168.0.1:1024",
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.output, test.input.String())
	}
}

func TestUnmarshal(t *testing.T) {
	var tests = map[string]struct {
		input   string
		output  []Peer
		failure bool
	}{
		"correctly parses peers": {
			input: string([]byte{127, 0, 0, 1, 0x00, 0x50, 1, 1, 1, 1, 0x01, 0xbb}),
			output: []Peer{
				{IP: net.IP{127, 0, 0, 1}, Port: 80},
				{IP: net.IP{1, 1, 1, 1}, Port: 443},
			},
			failure: false,
		},
		"not enough bytes in peers": {
			input:   string([]byte{127, 0, 0, 1, 0x00}),
			output:  nil,
			failure: true,
		},
	}
	for _, test := range tests {
		peer, err := Unmarshal([]byte(test.input))
		if test.failure {
			assert.NotNil(t, err) // expected error
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, peer)
	}
}
