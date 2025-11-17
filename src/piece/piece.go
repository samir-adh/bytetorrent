package piece

type PieceState int

const (
	Missing PieceState = iota
	HashError
	Downloaded
	Failed
)

type Piece struct {
	Index  int
	Hash   [20]byte
	Length int
}

type PieceResult struct {
	Index   int
	Payload []byte
	State PieceState
}
