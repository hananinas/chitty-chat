package server

import (
	"context"
	"log"
	"sync"

	"github.com/hananinas/chitty-chat/api"
	pb "github.com/hananinas/chitty-chat/api"
	"github.com/hananinas/chitty-chat/internal/chat"
	"google.golang.org/grpc"
)

func GetLamport() chat.LamportClock {
	return chat.LamportClock{}
}

func GetBroadcast() chat.BroadcastSubscription {
	return chat.BroadcastSubscription{}
}

// server is used to implement an unimplemented server.
type server struct {
	pb.UnimplementedChatServiceServer
	*Config
}

// cofniguration for the server
type Config struct {
	// keeps a map of all the clients that are connected to the server
	clients map[string]GetLamport();
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
			clients: make(map[string]),
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

// now i want to implement the methods of the server
// Join is a method that is called when a client wants to join the chat
// it takes a JoinRequest and returns a JoinResponse

func (s server) Join(ctx context.Context, req *api.JoinRequest) (*api.JoinResponse, error) {
	log.Printf("a client wants to join the chat")

	// if the client is not in the clients map, add it to the clients map
	// and return a JoinResponse with a status of OK
	log.Printf("Client %s joined", req.GetNodeName())
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	if _, ok := s.clients[req.GetNodeName()]; !ok {

	}

	return &api.JoinResponse{
		NodeId: req.GetNodeName(),
		Status: api.Status_OK,
		Lamport: &api.Lamport{
			NodeId: s.Name,
			Time:   s.lamport.GetTimestamp(),
		},
	}, nil
}
