package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	"github.com/jackpal/bencode-go"
	"github.com/ztrue/tracerr"
)

type BencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type BencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list"`
	Info         BencodeInfo `bencode:"info"`
}

func Open(r io.Reader) (*BencodeTorrent, error) {
	bto := BencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}

type TorrentFile struct {
	Announce    string     // the URL of the tracker
	InfoHash    [20]byte   // sha1 hash of the torrent file
	PiecesHash  [][20]byte // a hash list, i.e., a concatenation of each piece's SHA-1 hash.
	PieceLength int        // number of bytes per piece. This is commonly 2^8 KiB = 256 KiB = 262,144 B.
	Length      int        // size of the file in bytes
	Name        string     // suggested filename where the file is to be saved.
}

// Computes the info hash of the torrent
// i.e. the sha1 hash of the .torrent file.
func (bto *BencodeTorrent) InfoHash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, bto.Info)
	if err != nil {
		return [20]byte{}, err
	}
	hash := sha1.Sum(buf.Bytes())
	return hash, nil
}

// Takes the pieces field of the .torrent file and divides it
// into n 20 bytes chucks, the ith chunk corresponding to the sha1
// hash of the ith piece.
func (bto *BencodeTorrent) PiecesHash() ([][20]byte, error) {
	// TODO
	var piecesHash [][20]byte
	for i := 0; i < len(bto.Info.Pieces); i += 20 {
		var hash [20]byte
		copy(hash[:], bto.Info.Pieces[i:i+20])
		piecesHash = append(piecesHash, hash)
	}
	return piecesHash, nil
}

func (bto BencodeTorrent) ToTorrentFile() (TorrentFile, error) {
	var tor TorrentFile
	if bto.Announce != "" {
		tor.Announce = bto.Announce
	} else if len(bto.AnnounceList) > 0 {
		tor.Announce = bto.AnnounceList[0][0]
	} else {
		return tor, fmt.Errorf("no announce or announce-list found")
	}
	var err error
	tor.InfoHash, err = bto.InfoHash()
	if err != nil {
		return tor, tracerr.Wrap(err)
	}
	tor.PiecesHash, err = bto.PiecesHash()
	if err != nil {
		return tor, tracerr.Wrap(err)
	}
	tor.PieceLength = bto.Info.PieceLength
	tor.Length = bto.Info.Length
	tor.Name = bto.Info.Name
	return tor, nil
}

func OpenTorrentFile(filepath string) (*TorrentFile, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer file.Close()
	var bt BencodeTorrent
	err = bencode.Unmarshal(file, &bt)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	tf, err := bt.ToTorrentFile()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &tf, nil
}

func (tf *TorrentFile) getPieceBounds(index int) (int, int) {
	start := tf.PieceLength * index
	end := start + tf.PieceLength
	if end < tf.Length {
		return start, end
	} else {
		return start, tf.Length
	}
}

func (tf *TorrentFile) GetPieceLength(index int) int {
	start, end := tf.getPieceBounds(index)
	return end - start
}
