package torrentfile

import (
	"bittorrent-client-go/peers"
	"github.com/google/go-querystring/query"
	"github.com/jackpal/bencode-go"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// trackerResponse is the tracker response with "text/plain" document consisting of a bencoded dictionary
type trackerResponse struct {
	Failure     string `bencode:"failure reason"`  // if present, then no other keys may be present
	Warning     string `bencode:"warning message"` // optional, new
	Interval    int    `bencode:"interval"`        // in seconds
	MinInterval int    `bencode:"min interval"`    // optional, minimum announce interval
	TrackerID   string `bencode:"tracker id"`      // a string that the client should send back on its next announcements
	Complete    int    `bencode:"complete"`        // number of peers with the entire file, i.e. seeders
	Incomplete  int    `bencode:"incomplete"`      // number of non-seeder peers, aka "leechers"
	Peers       string `bencode:"peers"`           // list of peers (binary model), TODO: only binary model is supported
}

// trackerRequest is the parameters used in the client->tracker GET request
type trackerRequest struct {
	InfoHash   []string `url:"info_hash"`
	PeerID     []string `url:"peer_id"`
	Port       []string `url:"port"`
	Uploaded   []string `url:"uploaded"`
	Downloaded []string `url:"downloaded"`
	Left       []string `url:"left"`
	Compact    []string `url:"compact"`              // for binary model `compact` is set to 1
	NoPeerID   []string `url:"no_peer_id,omitempty"` // this option is ignored if `compact` is enabled
	Event      []string `url:"event,omitempty"`      // if specified, must be one of started, completed, stopped, (or empty which is the same as not being specified)
	IP         []string `url:"ip,omitempty"`         // optional, the true IP of the client
	NumWant    []string `url:"numwant,omitempty"`    // optional, number of peers that the client would like to receive from the tracker
	Key        []string `url:"key,omitempty"`        // optional, an additional identification that is not shared with any other peers
	TrackerID  []string `url:"trackerid,omitempty"`  // optional, if a previous `announce` contained a tracker id it should be set here
}

// buildTrackerURL builds initial tracker URL
func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := trackerRequest{
		InfoHash:   []string{string(t.InfoHash[:])},
		PeerID:     []string{string(peerID[:])},
		Port:       []string{strconv.Itoa(int(port))},
		Uploaded:   []string{"0"},
		Downloaded: []string{"0"},
		Left:       []string{strconv.Itoa(t.Length)},
		Compact:    []string{"1"}, // currently set to 1
	}
	request, err := query.Values(params)
	if err != nil {
		return "", err
	}
	base.RawQuery = (request).Encode()
	return base.String(), nil
}

// requestPeers requests a list of peers from tracker
func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	trackerURL, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}
	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(trackerURL)
	if err != nil {
		return nil, err
	}
	defer func(resp *http.Response) {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}(resp)
	trackerResp := trackerResponse{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}
	return peers.Unmarshal([]byte(trackerResp.Peers))
}
