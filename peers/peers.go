package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

// Peer encodes connection information for a peer
type Peer struct {
	IP   net.IP // TODO: currently only IPv4 is supported
	Port uint16
}

// Unmarshal parses peer IP addresses and ports from a buffer
func Unmarshal(peersBin []byte) ([]Peer, error) {
	const peerSize = 6 // 4 for IPv4, 2 for port
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		err := fmt.Errorf("received malformed peers")
		return nil, err
	}
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i = i + 1 {
		offset := peerSize * i
		peers[i].IP = peersBin[offset : offset+4]
		peers[i].Port = binary.BigEndian.Uint16(peersBin[offset+4 : offset+6])
	}
	return peers, nil
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
