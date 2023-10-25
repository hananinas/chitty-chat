package main

import (
	"log"
	"net"

	"github.com/hananinas/chitty-chat/internal/server"
)

const (
	port = ":4321"
	name = "Chit-chat"
)

func main() {
	log.Printf("starting chat-server on port %s", port)

	grpcServer, err := server.NewGrpcServer(name)
	if err != nil {
		log.Fatalf("could not create server: %v", err)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}

	defer func() {
		if err := lis.Close(); err != nil {
			log.Fatalf("could not close listener: %v", err)
		}
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("could not serve: %v", err)
	}

	log.Printf("Server started %s", name)

}
