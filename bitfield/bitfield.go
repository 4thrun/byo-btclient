package bitfield

// "bitfield" is one type of message,
// which is a data structure that peers use to efficiently encode which pieces they are able to send.
// A bitfield looks like a byte array, and to check which pieces they have,
// we just need to look at the positions of the bits set to 1.
// You can think of it like the digital equivalent of a coffee shop loyalty card.
// We start with a blank card of all 0, and flip bits to 1 to mark their positions as “stamped.”

// Bitfield represents the pieces that a peer has
type Bitfield []byte

// HasPiece tells if a bitfield has a particular index set
// returns true only when 0bxxxxxxx1 &　0b00000001 == 0b00000001
func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}
	offset := index % 8
	return bf[byteIndex]>>uint(7-offset)&1 != 0 // move the specified bit to the last
}

// SetPiece sets a bit in the bitfield
func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	if byteIndex < 0 || byteIndex >= len(bf) { // silently discard invalid bounded index
		return
	}
	offset := index % 8
	bf[byteIndex] |= 1 << uint(7-offset)
}
