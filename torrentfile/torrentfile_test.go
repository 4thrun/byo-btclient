package torrentfile

import (
	"encoding/json"
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var update = flag.Bool("update", false, "update .golden.json files")

func TestOpen(t *testing.T) {
	var torrent, err = Open("testdata/debian-edu-12.6.0-amd64-netinst.iso.torrent") // note that some .torrent files may not contain an `sannounce`
	require.Nil(t, err)
	var goldenPath = "testdata/debian-edu-12.6.0-amd64-netinst.iso.torrent.golden.json"
	if !*update { // switch
		serialized, err := json.MarshalIndent(torrent, "", "  ")
		require.Nil(t, err)
		_ = os.WriteFile(goldenPath, serialized, 0644)
	}
	var expected = TorrentFile{}
	golden, err := os.ReadFile(goldenPath)
	require.Nil(t, err)
	err = json.Unmarshal(golden, &expected)
	require.Nil(t, err)
	assert.Equal(t, expected, torrent)
}

func TestBencodeTorrent_toTorrentFile(t *testing.T) {
	var tests = map[string]struct {
		input   *BencodeTorrent
		output  TorrentFile
		failure bool
	}{
		"correct conversion": {
			input: &BencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Pieces:      "1234567890abcdefghijabcdefghij1234567890",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output: TorrentFile{
				Announce: "http://bttracker.debian.org:6969/announce",
				InfoHash: [20]byte{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
				PieceHashes: [][20]byte{
					{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
					{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				},
				PieceLength: 262144,
				Length:      351272960,
				Name:        "debian-10.2.0-amd64-netinst.iso",
			},
			failure: false,
		},
		"not enough bytes in pieces": {
			input: &BencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Pieces:      "1234567890abcdefghijabcdef", // only 26 bytes
					PieceLength: 262144,
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output:  TorrentFile{},
			failure: true,
		},
	}
	for _, test := range tests {
		to, err := test.input.toTorrentFile()
		if test.failure {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, to)
	}
}
