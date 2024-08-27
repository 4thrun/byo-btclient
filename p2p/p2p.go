package p2p

import (
	"bittorrent-client-go/client"
	"bittorrent-client-go/message"
	"bittorrent-client-go/peers"
	"bittorrent-client-go/torrentfile"
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize int = 16384

// MaxBlockLog is the number of unfulfilled requests a client can have in its pipeline
const MaxBlockLog int = 5

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers  []peers.Peer
	PeerID [20]byte
	File   torrentfile.TorrentFile
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlogged int
}

// Download downloads the .torrent and stores the entire file in memory
// TODO: do not store the entire file in memory
func (t *Torrent) Download() ([]byte, error) {
	log.Printf("starting download for %s\n", t.File.Name)
	workQueue := make(chan *pieceWork, len(t.File.PieceHashes)) // initialize queues for workers to retrieve work and send results
	results := make(chan *pieceResult)
	for index, hash := range t.File.PieceHashes {
		var length = t.calculatePieceSize(index)
		workQueue <- &pieceWork{
			index:  index,
			hash:   hash,
			length: length,
		}
	}
	defer close(workQueue)
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}
	var buf = make([]byte, t.File.Length)
	donePieces := 0
	for donePieces < len(t.File.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces += 1
		percent := float64(donePieces) / float64(len(t.File.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	return buf, nil
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.File.PieceLength
	end = begin + t.File.PieceLength
	if end > t.File.Length {
		end = t.File.Length
	}
	return begin, end
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := client.New(peer, t.PeerID, t.File.InfoHash)
	if err != nil {
		log.Printf("could not complete handshake with %s, disconnecting...\n", peer.IP)
		return
	}
	defer func(Conn net.Conn) {
		_ = Conn.Close()
		return
	}(c.Conn)
	log.Printf("completed handshake with %s\n", peer.IP)
	_ = c.SendUnchoke()
	_ = c.SendInterested()
	c.ClientInfo.AmInterested = true
	c.ClientInfo.AmChoking = false
	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw // put piece back to the queue
			continue
		}
		var buf, err = attemptDownloadPiece(c, pw)
		if err != nil {
			log.Printf("exiting on error: %s\n", err)
			workQueue <- pw // put piece back to the queue
			return
		}
		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("iece #%d failed integrity check\n", pw.index)
			workQueue <- pw
			continue
		}
		_ = c.SendHave(pw.index)
		results <- &pieceResult{index: pw.index, buf: buf}
	}
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	var hash = sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("index #%d failed integrity check\n", pw.index)
	}
	return nil
}

func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}
	_ = c.Conn.SetDeadline(time.Now().Add(30 * time.Second)) // setting a deadline helps get unresponsive peers unstuck
	defer func(Conn net.Conn) {                              // 30 seconds is more than enough time to download a 262KB piece
		_ = Conn.Close()
		return
	}(c.Conn)
	for state.downloaded < pw.length {
		if !state.client.ClientInfo.AmChoking { // if choked, send requests until we have enough unfulfilled requests
			for state.backlogged < MaxBlockLog && state.requested < pw.length {
				var blockSize = MaxBlockSize
				if pw.length-state.requested < blockSize { // last block might be shorter than the typical block
					blockSize = pw.length - state.requested
				}
				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				} else {
					state.backlogged++
					state.requested += blockSize
				}
			}
		}
		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

// readMessage validates a message format
func (state *pieceProgress) readMessage() error {
	var msg, err = state.client.Read() // this call blocks
	if err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("empty message")
	}
	if msg.Prefix == uint32(0) { // keep-alive message
		return nil
	}
	switch msg.MessageID {
	case message.MsgUnchoke:
		state.client.ClientInfo.AmChoking = false
	case message.MsgChoke:
		state.client.ClientInfo.AmChoking = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, msg, state.buf)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlogged--
	}
	return nil
}
