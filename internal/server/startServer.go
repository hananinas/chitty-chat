package server

import (
	"log"
	"net"
)

const (
	port = ":4321"
	name = "Chit-chat"
)

func StartServer() {
	log.Printf("starting chat-server on port %s", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}
	grpcServer, err := NewGrpcServer(name)
	if err != nil {
		log.Fatalf("could not create server: %v", err)
	}

	defer func() {
		if err := lis.Close(); err != nil {
			log.Fatalf("could not close listener: %v", err)
		}
	}()
	log.Printf("Server started!")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("could not serve: %v", err)
	}

}
