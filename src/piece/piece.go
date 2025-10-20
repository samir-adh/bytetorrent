package piece

type PieceState int

const (
	Missing PieceState = iota
	Downloading
	Downloaded
)

type Piece struct {
	Index int
	State PieceState
	Hash  [20]byte
	Length int
}