package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"log"
	"os"
)

// bencodeInfo represents Info Dictionary in Single File Mode
type bencodeInfo struct { // TODO: Multiple File Mode to be supported
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	//Private     int    `bencode:"private"` // optional, for PT (Private Tracker)
	Name   string `bencode:"name"`
	Length int    `bencode:"length"`
	//MD5Sum      string `bencode:"md5sum"` // optional
}

// BencodeTorrent represents Bencode format (.torrent format)
type BencodeTorrent struct {
	Info         bencodeInfo `bencode:"info"`
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list"` // optional
	Date         int64       `bencode:"creation date"` // optional, in standard UNIX epoch format
	Comment      string      `bencode:"comment"`       // optional
	CreatedBy    string      `bencode:"created by"`    // optional
	Encoding     string      `bencode:"encoding"`      // optional
}

// TorrentFile represents actual file information we need
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte // SHA-1 hash of the entire bencoded `info` dict
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

// hash returns SHA-1 of the bencoded `info` dict
func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	} else {
		return sha1.Sum(buf.Bytes()), nil
	}
}

// splitHashes returns a list of hashes of pieces
func (i *bencodeInfo) splitHashes() ([][20]byte, error) {
	var length = 20 // length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%length != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(buf))
	}
	index := len(buf) / length
	hashes := make([][20]byte, index)
	for i := 0; i < index; i++ {
		_ = copy(hashes[i][:], buf[i*length:(i+1)*length])
	}
	return hashes, nil
}

// toTorrentFile saves useful fields in BencodeTorrent into TorrentFile
func (bto *BencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	return TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}, nil
}

// Open parses a torrent file
func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Printf("error closing file: %v", file)
		}
	}(file)
	var bto BencodeTorrent
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}
	return bto.toTorrentFile()
}

// DownloadToFile downloads a torrent and writes it to a file
func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
<<<<<<< HEAD
	_, err := rand.Read(peerID[:]) // randomly generated peer ID
	if err != nil {
		return err
	}
	peers, err := t.requestPeers(peerID, Port)
	if err != nil {
		return err
	}
	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		Name:        t.Name,
		Length:      t.Length,
		PieceLength: t.PieceLength,
	}
	buf, err := torrent.Download()
	if err != nil {
		return err
	}
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(output *os.File) {
		_ = output.Close()
	}(output)
	_, err = output.Write(buf)
=======
	_, err := rand.Read(peerID[:]) // randomly generated
>>>>>>> parent of 910f03c (test torrentfile)
	if err != nil {
		return err
	}
	// TODO: to be implemented later
	return nil
}
