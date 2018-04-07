package main

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Session struct {
	masterConn net.Conn
	*rpc.Client
}

func NewSession(masterConn, clientConn net.Conn) *Session {
	return &Session{
		masterConn: masterConn,
		Client:     rpc.NewClientWithCodec(jsonrpc.NewClientCodec(clientConn)),
	}
}

func (s *Session) Close() {
	s.Client.Close()
	s.masterConn.Close()
}
