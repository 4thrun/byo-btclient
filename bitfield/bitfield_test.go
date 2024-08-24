package bitfield

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBitfield_HasPiece(t *testing.T) {
	bf := Bitfield{0b01010100, 0b01010100}
	expected := []bool{
		false, true, false, true, false, true, false, false,
		false, true, false, true, false, true, false, false,
		false, false, false, false,
	}
	for i := 0; i < len(expected); i = i + 1 {
		assert.Equal(t, expected[i], bf.HasPiece(i))
	}
}

func TestBitfield_SetPiece(t *testing.T) {
	var tests = []struct {
		input  Bitfield
		index  int
		output Bitfield
	}{
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  4,
			output: Bitfield{0b01011100, 0b01010100},
		},
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  9,
			output: Bitfield{0b01010100, 0b01010100},
		},
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  15,
			output: Bitfield{0b01010100, 0b01010101},
		},
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  19,
			output: Bitfield{0b01010100, 0b01010100},
		},
	}
	for _, test := range tests {
		bf := test.input
		bf.SetPiece(test.index)
		assert.Equal(t, test.output, bf)
	}
}
