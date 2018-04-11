package main

import (
	"encoding/binary"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/yamux"
)

type Server struct {
	clientid  int32
	sessions  sync.Map // map[id]*Session
	rpcServer *rpc.Server

	playerCallback func(string, int32)
}

func NewServer() *Server {
	return &Server{
		rpcServer: rpc.NewServer(),
	}
}

func (s *Server) serveRpc(sess *yamux.Session) {
	conn, err := sess.Accept()
	if err != nil {
		log.Print(err)
		return
	}
	s.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	id := atomic.AddInt32(&s.clientid, 1)
	log.Printf("allocated %d for %s", id, conn.RemoteAddr())
	// send id to client, handshake done.
	binary.Write(conn, binary.BigEndian, id)

	sess, err := yamux.Server(conn, nil)
	if err != nil {
		log.Print(err)
		return
	}

	clientConn, err := sess.Open()
	if err != nil {
		log.Print(err)
		return
	}
	session := NewSession(conn, clientConn)
	s.sessions.Store(id, session)
	s.playerCallback("online", id)
	s.serveRpc(sess)
	s.sessions.Delete(id)
	s.playerCallback("offline", id)
	log.Printf("%s(%d) closed connection", conn.RemoteAddr(), id)
}

func (s *Server) RegisterService(name string, service interface{}) error {
	return s.rpcServer.RegisterName(name, service)
}

func (s *Server) RangeSession(f func(id int32, sess *Session)) {
	s.sessions.Range(func(k, v interface{}) bool {
		f(k.(int32), v.(*Session))
		return true
	})
}

func (s *Server) SetPlayerCallback(callback func(string, int32)) {
	s.playerCallback = callback
}

func (s *Server) Serve(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go s.handleConn(conn)
	}
}
