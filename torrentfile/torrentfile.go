package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
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
	err := bencode.Marshal(&buf, &bto.Info)
	if err != nil {
		return [20]byte{}, err
	}
	hash := sha1.Sum(buf.Bytes())
	return hash, nil
}



func (bto *bencodeTorrent) PiecesHash() ([][20]byte, error) {
	// TODO
	return [][20]byte{}, nil
}

func (bto bencodeTorrent) ToTorrentFile() (TorrentFile, error) {
	var tor TorrentFile
	tor.Announce = bto.Announce
	var err error
	tor.InfoHash, err = bto.InfoHash()
	if err != nil {
		return tor, fmt.Errorf("failed to get info hash: %v", err)
	}
	tor.PiecesHash, err = bto.PiecesHash()
	if err != nil {
		return tor, err
	}
	tor.PieceLength = bto.Info.PieceLength
	tor.Length = bto.Info.Length
	tor.Name = bto.Info.Name
	return tor, nil
}


func OpenTorrentFile(filepath string) (*TorrentFile, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var bt bencodeTorrent
	err = bencode.Unmarshal(file,&bt)
	if err != nil {
		return nil, err
	}
	tf,err := bt.ToTorrentFile()
	if err != nil {
		return nil, fmt.Errorf("failed to convert 'bencodeTorrent' type to 'TorrentFile' type: %v", err)
	}
	return &tf, nil
}



