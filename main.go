package main

import (
	"flag"
	"log"
	"net"
)

var (
	listenAddr = flag.String("l", ":8421", "listen address")
)

func main() {
	flag.Parse()

	err := InitStore()
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer()
	blockService := NewBlockService(server)
	playerService := NewPlayerService(server)
	server.RegisterService("Block", blockService)
	server.RegisterService("Player", playerService)
	server.Serve(l)
}
