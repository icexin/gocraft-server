package gocraft

import (
	"encoding/binary"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/hashicorp/yamux"
)

type Client struct {
	rpcServer *rpc.Server

	ClientId int32
	*rpc.Client
}

func NewClient() *Client {
	return &Client{
		rpcServer: rpc.NewServer(),
	}
}

func (c *Client) doServer(sess *yamux.Session) {
	clientConn, err := sess.Accept()
	if err != nil {
		log.Panic(err)
		return
	}
	c.rpcServer.ServeCodec(jsonrpc.NewServerCodec(clientConn))
}

func (c *Client) doClient(sess *yamux.Session) {
	clientConn, err := sess.Open()
	if err != nil {
		log.Panic(err)
		return
	}
	c.Client = rpc.NewClientWithCodec(jsonrpc.NewClientCodec(clientConn))
}

func (c *Client) Start(conn net.Conn) {
	binary.Read(conn, binary.BigEndian, &c.ClientId)

	sess, err := yamux.Client(conn, nil)
	if err != nil {
		log.Panic(err)
	}

	go c.doServer(sess)
	c.doClient(sess)
}

func (c *Client) RegisterService(name string, service interface{}) error {
	return c.rpcServer.RegisterName(name, service)
}

func (c *Client) Close() {
	c.Client.Close()
}
