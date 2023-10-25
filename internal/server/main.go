package server

import (
	"log"
	"sync"

	"github.com/hananinas/chitty-chat/api"
	pb "github.com/hananinas/chitty-chat/api"
	"github.com/hananinas/chitty-chat/internal/chat"
	"google.golang.org/grpc"
)

type BroadcastSubscription struct {
	stream   api.ChatService_BroadcastServer
	finished chan<- bool
}

// server is used to implement an unimplemented server.
type server struct {
	pb.UnimplementedChatServiceServer
	*Config
}

// cofniguration for the server
type Config struct {
	// keeps a map of all the clients that are connected to the server
	clients map[string]BroadcastSubscription
	// mutex to lock the clients map
	clientsMu sync.Mutex
	// a clock to keep track of the lamport timestamp of the server

	lamport chat.LamportClock
	// the name of the server
	Name string
}

// NewServer creates a new server
func NewServer(name string) (*server, error) {
	chatServer := server{
		Config: &Config{
			clients: make(map[string]BroadcastSubscription),
			Name:    name,
			lamport: chat.LamportClock{Node: name},
		},
	}
	return &chatServer, nil
}

// NewGrpcServer creates a new gRPC server and registers the ChatServiceServer
func NewGrpcServer(name string) (*grpc.Server, error) {
	grpcServer := grpc.NewServer()
	s, err := NewServer(name)

	if err != nil {
		return nil, err
	}

	api.RegisterChatServiceServer(grpcServer, *s)
	log.Printf("Starting server %s", name)
	return grpcServer, nil
}
