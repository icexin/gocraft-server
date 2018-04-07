package main

import (
	"log"
	"sync"

	"github.com/icexin/gocraft-server/proto"
)

type BlockService struct {
	mutex  sync.Mutex
	server *Server
}

func NewBlockService(s *Server) *BlockService {
	return &BlockService{
		server: s,
	}
}

func (s *BlockService) UpdateBlock(req *proto.UpdateBlockRequest, rep *proto.UpdateBlockResponse) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	log.Printf("UpdateBlock: %v", req)
	version := GenrateChunkVersion()
	store.UpdateBlock(Vec3{req.X, req.Y, req.Z}, req.W)
	store.UpdateChunkVersion(Vec3{req.P, 0, req.Q}, version)
	req.Version = version
	rep.Version = version
	s.server.RangeSession(func(id int32, sess *Session) {
		if id == req.Id {
			return
		}
		sess.Go("Block.UpdateBlock", req, new(proto.UpdateBlockResponse), nil)
	})
	return nil
}

func (s *BlockService) FetchChunk(req *proto.FetchChunkRequest, rep *proto.FetchChunkResponse) error {
	id := Vec3{req.P, 0, req.Q}
	version := store.GetChunkVersion(id)
	rep.Version = version
	if req.Version == version {
		return nil
	}
	store.RangeBlocks(id, func(bid Vec3, w int) {
		rep.Blocks = append(rep.Blocks, [...]int{bid.X, bid.Y, bid.Z, w})

	})
	return nil
}

type PlayerService struct {
	mutex   sync.Mutex
	server  *Server
	players map[int32]proto.PlayerState
}

func NewPlayerService(server *Server) *PlayerService {
	s := &PlayerService{
		server:  server,
		players: make(map[int32]proto.PlayerState),
	}
	server.SetPlayerCallback(s.onPlayerCallback)
	return s
}

func (s *PlayerService) UpdateState(req *proto.UpdateStateRequest, rep *proto.UpdateStateResponse) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.players[req.Id]; !ok {
		return nil
	}
	s.players[req.Id] = req.State
	rep.Players = make(map[int32]proto.PlayerState)
	for id, state := range s.players {
		if id == req.Id {
			continue
		}
		rep.Players[id] = state
	}
	return nil
}

func (s *PlayerService) onPlayerCallback(action string, id int32) {
	switch action {
	case "online":
		s.addPlayer(id)
	case "offline":
		s.removePlayer(id)
	}
}

func (s *PlayerService) removePlayer(pid int32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.players, pid)
	req := &proto.RemovePlayerRequest{
		Id: pid,
	}
	s.server.RangeSession(func(id int32, sess *Session) {
		if id == pid {
			return
		}
		sess.Go("Player.RemovePlayer", req, new(proto.RemovePlayerResponse), nil)
	})
}

func (s *PlayerService) addPlayer(pid int32) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.players[pid] = proto.PlayerState{}
}
