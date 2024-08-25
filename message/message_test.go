package message

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage_String(t *testing.T) {
	var tests = []struct {
		input  *Message
		output string
	}{
		{input: nil, output: "malformed message"},
		{input: &Message{Prefix: uint32(0)}, output: "keep-alive"},
		{input: &Message{Prefix: uint32(1), MessageID: MsgChoke}, output: "<message ID> 0 (choke), <payload> length 0"},
		{input: &Message{Prefix: uint32(1), MessageID: MsgUnchoke}, output: "<message ID> 1 (unchoke), <payload> length 0"},
		{input: &Message{Prefix: uint32(1), MessageID: MsgInterested}, output: "<message ID> 2 (interested), <payload> length 0"},
		{input: &Message{Prefix: uint32(1), MessageID: MsgNotInterested}, output: "<message ID> 3 (not interested), <payload> length 0"},
		{input: &Message{Prefix: uint32(1 + 4), MessageID: MsgHave, Payload: []byte{9, 8, 7, 6}}, output: "<message ID> 4 (have), <payload> length 4"},
		{input: &Message{Prefix: uint32(1 + 2), MessageID: MsgBitfield, Payload: []byte{0b10110011, 0b01001100}}, output: "<message ID> 5 (bitfield), <payload> length 2"},
		{input: &Message{Prefix: uint32(1 + 4 + 4 + 4), MessageID: MsgRequest, Payload: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}}, output: "<message ID> 6 (request), <payload> length 12"},
		{input: &Message{Prefix: uint32(1 + 4 + 4 + 1), MessageID: MsgPiece, Payload: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1}}, output: "<message ID> 7 (piece), <payload> length 9"},
		{input: &Message{Prefix: uint32(1 + 4 + 4 + 4), MessageID: MsgCancel, Payload: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}}, output: "<message ID> 8 (cancel), <payload> length 12"},
		{input: &Message{Prefix: uint32(1 + 2), MessageID: MsgPort, Payload: []byte{12, 34}}, output: "<message ID> 9 (port), <payload> length 2"},
		{input: &Message{Prefix: uint32(99), MessageID: 88, Payload: []byte{123, 32}}, output: "<message ID> 88 (unknown (<id=88>)), <payload> length 2"},
	}
	for _, test := range tests {
		assert.Equal(t, test.output, test.input.String())
	}
}

func TestRead(t *testing.T) {
	tests := map[string]struct {
		failure bool
		input   []byte
		output  *Message
	}{
		"normal message": {
			input:   []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
			output:  &Message{MessageID: MsgHave, Payload: []byte{1, 2, 3, 4}, Prefix: uint32(5)},
			failure: false,
		},
		"keep-alive": {
			input:   []byte{0, 0, 0, 0},
			output:  &Message{Prefix: uint32(0)},
			failure: false,
		},
		"length too short": {
			input:   []byte{1, 2, 3},
			output:  nil,
			failure: true,
		},
		"buffer too short for length": {
			input:   []byte{0, 0, 0, 5, 4, 1, 2},
			output:  nil,
			failure: true,
		},
	}
	for _, test := range tests {
		var reader = bytes.NewReader(test.input)
		m, err := Read(reader)
		if test.failure {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, m, test.output)
	}
}

func TestFormatRequest(t *testing.T) {
	msg := FormatRequest(4, 567, 4321)
	expected := &Message{
		MessageID: MsgRequest,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x04, // <index>
			0x00, 0x00, 0x02, 0x37, // <begin>
			0x00, 0x00, 0x10, 0xe1, // <length>
		},
		Prefix: uint32(12 + 1),
	}
	assert.Equal(t, expected, msg)
}

func TestFormatHave(t *testing.T) {
	msg := FormatHave(4)
	expected := &Message{
		Prefix:    uint32(1 + 4),
		MessageID: MsgHave,
		Payload:   []byte{0x00, 0x00, 0x00, 0x04},
	}
	assert.Equal(t, expected, msg)
}

func TestMessage_Serialize(t *testing.T) {
	var tests = map[string]struct {
		input  *Message
		output []byte
	}{
		"other message": {
			input:  &Message{MessageID: MsgHave, Payload: []byte{1, 2, 3, 4}, Prefix: uint32(5)},
			output: []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
		},
		"keep-alive": {
			input:  &Message{Prefix: 0},
			output: []byte{0, 0, 0, 0},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.output, test.input.Serialize())
	}
}

func TestParseHave(t *testing.T) {
	var tests = map[string]struct {
		input   *Message
		output  int
		failure bool
	}{
		"valid message": {
			input:   &Message{MessageID: MsgHave, Payload: []byte{0x00, 0x00, 0x00, 0x04}, Prefix: uint32(5)},
			output:  4,
			failure: false,
		},
		"wrong message type": {
			input:   &Message{MessageID: MsgPiece, Payload: []byte{0x00, 0x00, 0x00, 0x04}, Prefix: uint32(5)},
			output:  0,
			failure: true,
		},
		"payload too short": {
			input:   &Message{MessageID: MsgHave, Payload: []byte{0x00, 0x00, 0x04}, Prefix: uint32(5)},
			output:  0,
			failure: true,
		},
		"payload too long": {
			input:   &Message{MessageID: MsgHave, Payload: []byte{0x00, 0x00, 0x00, 0x00, 0x04}, Prefix: uint32(5)},
			output:  0,
			failure: true,
		},
	}
	for _, test := range tests {
		var index, err = ParseHave(test.input)
		if test.failure {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, index)
	}
}

func TestParsePiece(t *testing.T) {
	var tests = map[string]struct {
		inputIndex int
		inputBuf   []byte
		inputMsg   *Message
		outputN    int
		outputBuf  []byte
		failure    bool
	}{
		"valid piece": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				MessageID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // <index>
					0x00, 0x00, 0x00, 0x02, // <begin>
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // <block>
				},
				Prefix: uint32(15),
			},
			outputBuf: []byte{0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x00},
			outputN:   6,
			failure:   false,
		},
		"wrong message type": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				MessageID: MsgChoke,
				Payload:   []byte{},
				Prefix:    uint32(1),
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			failure:   true,
		},
		"payload too short": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				MessageID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // <index>
					0x00, 0x00, 0x00, // malformed offset
				},
				Prefix: uint32(8),
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			failure:   true,
		},
		"wrong index": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				MessageID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x06, // <index> is 6, not 4
					0x00, 0x00, 0x00, 0x02, // <begin>
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // <block>
				},
				Prefix: uint32(15),
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			failure:   true,
		},
		"offset too high": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				MessageID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // <index> is 6, not 4
					0x00, 0x00, 0x00, 0x0c, // <begin> is 12 > 10
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // <block>
				},
				Prefix: uint32(15),
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			failure:   true,
		},
		"offset ok but payload too long": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				MessageID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // <index> is 6, not 4
					0x00, 0x00, 0x00, 0x02, // <begin> is ok
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x0a, 0x0b, 0x0c, 0x0d, // <block> is 10 long but begin=2 (too long for input buffer)
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			failure:   true,
		},
	}
	for _, test := range tests {
		n, err := ParsePiece(test.inputIndex, test.inputMsg, test.inputBuf)
		if test.failure {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.outputBuf, test.inputBuf)
		assert.Equal(t, test.outputN, n)
	}
}
