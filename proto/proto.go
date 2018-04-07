package proto

// block service

type UpdateBlockRequest struct {
	Id      int32
	P, Q    int
	X, Y, Z int
	W       int
	Version string // used by server
}

type UpdateBlockResponse struct {
	Version string
}

type FetchChunkRequest struct {
	P, Q    int
	Version string
}

type FetchChunkResponse struct {
	Blocks  [][4]int
	Version string
}

// player service

type PlayerState struct {
	X, Y, Z float32
	Rx, Ry  float32
}

type UpdateStateRequest struct {
	Id    int32
	State PlayerState
}

type UpdateStateResponse struct {
	Players map[int32]PlayerState
}

type RemovePlayerRequest struct {
	Id int32
}

type RemovePlayerResponse struct {
}
