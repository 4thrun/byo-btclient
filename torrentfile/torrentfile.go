package torrentfile

import (
	"io"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

// bencodeInfo represents Info Dictionary in Single File Mode
type bencodeInfo struct { // TODO: Multiple File Mode to be supported
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Private     int    `bencode:"private"` // optional, for PT (Private Tracker)
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	MD5Sum      string `bencode:"md5sum"` // optional
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
	Announce    []byte
	InfoHash    [20]byte // SHA-1 hash of the entire bencoded `info` dict
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        []byte
}

// TODO: to be implemented later
//func (bto *BencodeTorrent) toTorrentFile() (TorrentFile, error) {
//
//}

// buildTrackerURL builds initial tracker URL
func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(string(t.Announce))
	if err != nil {
		return "", err
	}
	params := url.Values{ // HTTP GET parameters
		"info_hash":  []string{string(t.InfoHash[:])}, // SHA-1 hash of the entire bencoded `info` dict
		"peer_id":    []string{string(peerID[:])},     // a 20-byte name to identify THIS client to trackers and peers
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.Itoa(t.Length)},
		"compact":    []string{"1"},
		// "no_peer_id": []string{}, // this option is ignored if `compact` is enabled
		// "event": []string{},
		// "ip": []string{}, // optional, the true IP of the client
		// "numwant": []string{}, // optional, the number of peers that th client would like to receive from the tracker
		// "key": []string{}, // optional, an additional identification that is not shared with any other peers
		// "trackerid": []string{}, // optional, if a previous `announce` contained a tracker id it should be set here
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
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
