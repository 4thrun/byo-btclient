package torrentfile

import (
	"io"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Length      int    `bencode:"length"`
	Name        []byte `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`
}

// BencodeTorrent represents Bencode format (.torrent format)
type BencodeTorrent struct {
	Announce []byte      `bencode:"announce"`
	Comment  []byte      `bencode:"comment"`
	Date     int         `bencode:"creation date"`
	Info     bencodeInfo `bencode:"info"`
}

// TorrentFile represents actual file information
type TorrentFile struct {
	Announce    []byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        []byte
}

// Open parses a torrent file
func Open(r io.Reader) (*BencodeTorrent, error) {
	bto := BencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}
