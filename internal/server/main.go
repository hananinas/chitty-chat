package server

import (
	"context"
	"errors"
	"fmt"
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
	// keeps A map of all the clients that are connected to the server
	Clients map[string]BroadcastSubscription
	// mutex to lock the clients map
	clientsMu sync.Mutex
	// A clock to keep track of the lamport timestamp of the server
	Name string

	lamport chat.LamportClock
	// the name of the server
}

func (s *server) GetLamport() uint32 {
	return s.Config.lamport.GetTimestamp()
}

func (s *server) getName() string {
	return s.Config.Name
}

func (s *server) CompLamport(other uint32) {
	s.lamport.CompOtherClock(other)
}

func (s *server) MoveLamport() {
	s.lamport.Move()
}

// NewServer creates A new server
func NewServer(name string) (*server, error) {
	chatServer := server{

		Config: &Config{
			Clients: map[string]BroadcastSubscription{},
			Name:    name,
			lamport: chat.LamportClock{Node: name},
		},
	}
	return &chatServer, nil
}

// NewGrpcServer creates A new gRPC server and registers the ChatServiceServer
func NewGrpcServer(name string) (*grpc.Server, error) {
	grpcServer := grpc.NewServer()
	s, err := NewServer(name)

	if err != nil {
		return nil, err
	}

	api.RegisterChatServiceServer(grpcServer, *s)
	log.Printf("Starting server %s", name)
	s.MoveLamport()
	return grpcServer, nil
}

// now i want to implement the methods of the server
// Join is A method that is called when A client wants to join the chat
// it takes A JoinRequest and returns A JoinResponse
func (s server) Join(ctx context.Context, req *api.JoinRequest) (*api.JoinResponse, error) {
	log.Printf("[Server: %s time: %d] Received A JOIN req from node %s", s.getName(), s.GetLamport(), req.Lamport.GetNodeId())
	s.CompLamport(req.Lamport.GetTime())
	// add the client to the broadcast
	err := s.addClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	// if the client is not in the clients map, add it to the clients map
	// and return A JoinResponse with A status of OK
	s.MoveLamport()
	s.broadcast(fmt.Sprintf("Client %s joined the chat server", req.GetNodeName()))

	return &api.JoinResponse{
		NodeId: req.GetNodeName(),
		Status: api.Status_OK,
		Lamport: &api.Lamport{
			NodeId: s.getName(),
			Time:   s.lamport.GetTimestamp(),
		},
	}, nil
}

func (s *server) addClient(id string) error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if _, ok := s.Clients[id]; !ok {
		s.Clients[id] = BroadcastSubscription{}
	} else {
		log.Printf("client with id %s already exists ", id)
		return errors.New("client already exists")
	}

	if id == "" {
		log.Printf("client has no id")

		return errors.New("client has no id")
	}

	log.Printf("[%s] Added new client %s", s.Name, id)

	return nil
}

//now that we have A client who has joined we want that client to be able to leave the broadcast

func (s server) Leave(ctx context.Context, req *api.LeaveRequest) (*api.LeaveResponse, error) {
	log.Printf("[Server: %s time: %d] -- Received A Leave req from node %s", s.getName(), s.GetLamport(), req.Lamport.GetNodeId())
	s.CompLamport(req.Lamport.GetTime())

	// remove the client from the broadcast
	err := s.removeClient(req.SenderId)
	log.Printf(req.SenderId)

	if err != nil {
		return nil, err
	}

	s.MoveLamport()
	s.broadcast(fmt.Sprintf("Client %s left the chat server", req.GetSenderId()))
	return &api.LeaveResponse{
		NodeId: req.GetSenderId(),
		Status: api.Status_OK,
		Lamport: &api.Lamport{
			NodeId: s.Name,
			Time:   s.lamport.GetTimestamp(),
		},
	}, nil
}

// remove the client from the broadcast
func (s server) removeClient(id string) error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if _, ok := s.Clients[id]; !ok {
		return errors.New("client does not exist")
	}

	delete(s.Clients, id)

	log.Printf("[%s] Removed client %s", s.Name, id)

	return nil
}

// publish response

func (s server) Send(ctx context.Context, req *api.Message) (*api.PublishResponse, error) {
	log.Printf("[Server: %s time: %d] -- Received A Publish req from node %s", s.getName(), s.GetLamport(), req.Lamport.GetNodeId())
	s.CompLamport(req.Lamport.GetTime())
	// send the client message to all the clients
	s.broadcast(fmt.Sprintf("%s: %s", req.GetLamport(), req.GetContent()))

	return &api.PublishResponse{
		MessageHash: "",
		Lamport: &api.Lamport{
			Time:   s.GetLamport(),
			NodeId: s.getName(),
		},
		Status: api.Status_OK,
	}, nil
}

func (s *server) broadcast(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	log.Printf("%d clients found", len(s.Clients))
	for id, sub := range s.Clients {
		if sub.stream != nil {
			log.Printf("[Server: %s time: %d] >>>  broadcasting message >>> %s", s.getName(), s.GetLamport(), msg)
			err := sub.stream.Send(&api.Message{
				Lamport: &api.Lamport{
					Time:   s.GetLamport(),
					NodeId: s.getName(),
				},
				Content: msg,
			})
			if err != nil {
				log.Printf("could not send message to client %s: %v", id, err)
			}
		}
	}
}

// scope

func (s server) Broadcast(req *api.BroadcastSubscription, bsv api.ChatService_BroadcastServer) error {
	log.Printf("[Server: %s time: %d] -- Received req from client %s wants to subscribe to broadcast", s.getName(), s.GetLamport(), req.GetReceiver())

	fin := make(chan bool)

	err := s.addBroadcastChannelToClient(req.GetReceiver(), bsv, fin)
	if err != nil {
		return err
	}

	<-fin
	return nil
}

func (s *server) addBroadcastChannelToClient(id string, cs api.ChatService_BroadcastServer, fin chan bool) error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if _, ok := s.Clients[id]; !ok {
		log.Printf("client with id %s does not exist ", id)
		return errors.New("client does not exist")
	}

	s.Clients[id] = BroadcastSubscription{
		stream:   cs,
		finished: fin,
	}
	log.Printf("[Server: %s time: %d] -- Client %s finished broadcast subscription ", s.getName(), s.GetLamport(), id)

	return nil
}
