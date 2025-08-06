package torrentFile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	"github.com/jackpal/bencode-go"
	"github.com/ztrue/tracerr"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list"`
	Info         bencodeInfo `bencode:"info"`
}

func Open(r io.Reader) (*bencodeTorrent, error) {
	bto := bencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PiecesHash  [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (bto *bencodeTorrent) InfoHash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, bto.Info)
	if err != nil {
		return [20]byte{}, err
	}
	hash := sha1.Sum(buf.Bytes())
	return hash, nil
}

func (bto *bencodeTorrent) PiecesHash() ([][20]byte, error) {
	// TODO
	var piecesHash [][20]byte
	for i := 0; i < len(bto.Info.Pieces); i += 20 {
		var hash [20]byte
		copy(hash[:], bto.Info.Pieces[i:i+20])
		piecesHash = append(piecesHash, hash)
	}
	return piecesHash, nil
}

func (bto bencodeTorrent) ToTorrentFile() (TorrentFile, error) {
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
	var bt bencodeTorrent
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
