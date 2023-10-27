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
	// keeps a map of all the clients that are connected to the server
	Clients map[string]BroadcastSubscription
	// mutex to lock the clients map
	clientsMu sync.Mutex
	// a clock to keep track of the lamport timestamp of the server
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

func (s *server) BumpLamport(other uint32) {
	s.lamport.CompOtherClock(other)
}

func (s *server) TickLamport() {
	s.lamport.Move()
}

// NewServer creates a new server
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

// NewGrpcServer creates a new gRPC server and registers the ChatServiceServer
func NewGrpcServer(name string) (*grpc.Server, error) {
	grpcServer := grpc.NewServer()
	s, err := NewServer(name)

	if err != nil {
		return nil, err
	}

	api.RegisterChatServiceServer(grpcServer, *s)
	log.Printf("Starting server %s", name)
	s.TickLamport()
	return grpcServer, nil
}

// now i want to implement the methods of the server
// Join is a method that is called when a client wants to join the chat
// it takes a JoinRequest and returns a JoinResponse
// i think we should move this maybe
func (s server) Join(ctx context.Context, req *api.JoinRequest) (*api.JoinResponse, error) {
	log.Printf("a client wants to join the chat")

	// add the client to the broadcast
	err := s.addClient(req.GetNodeName())
	if err != nil {
		return nil, err
	}
	// if the client is not in the clients map, add it to the clients map
	// and return a JoinResponse with a status of OK
	log.Printf("Client %s joined the broadcast", req.GetNodeName())
	s.TickLamport()
	return &api.JoinResponse{
		NodeId: req.GetNodeName(),
		Status: api.Status_OK,
		Lamport: &api.Lamport{
			NodeId: s.Name,
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

//now that we have a client who has joined we want that client to be able to leave the broadcast

func (s server) Leave(ctx context.Context, req *api.LeaveRequest) (*api.LeaveResponse, error) {
	log.Printf("a client wants to leave the chat")
	s.TickLamport()

	// remove the client from the broadcast
	err := s.removeClient(req.GetSenderId())

	if err != nil {
		return nil, err
	}
	// if the client is not in the clients map, add it to the clients map
	// and return a JoinResponse with a status of OK
	log.Printf("Client %s left the broadcast", req.GetSenderId())
	fmt.Printf("%s left chitty chat", s.Name)
	return &api.LeaveResponse{
		NodeId: req.GetSenderId(),
		Status: api.Status_OK,
		Lamport: &api.Lamport{
			NodeId: s.Name,
		},
	}, nil
}

// remove the client from the broadcast
func (s server) removeClient(id string) error {
	if _, ok := s.Clients[id]; ok {
		delete(s.Clients, id)
	} else {
		log.Printf("client with id %s does not exist ", id)
	}
	return nil
}

// publish response

func (s server) Send(ctx context.Context, req *api.Message) (*api.PublishResponse, error) {
	s.TickLamport()

	// remove the client from the broadcast
	s.clientSend(fmt.Sprintf("%s: %s", req.GetLamport(), req.GetContent()))

	return &api.PublishResponse{
		MessageHash: "",
		Lamport: &api.Lamport{
			Time:   s.GetLamport(),
			NodeId: s.getName(),
		},
		Status: api.Status_OK,
	}, nil
}

func (s *server) clientSend(msg string) {
	log.Printf("Ready to broadcast message %s", msg)

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	clients := s.Clients
	var wg sync.WaitGroup
	wg.Add(len(clients))
	for _, ss := range clients {
		go func(ss BroadcastSubscription) {
			defer wg.Done()
			log.Printf("Broadcasting message %s ", msg)
			if ss.stream != nil {
				err := ss.stream.Send(&api.Message{
					Lamport: &api.Lamport{
						Time:   s.GetLamport(),
						NodeId: s.getName(),
					},
					Content: msg,
				})
				if err != nil {
					log.Printf("Error, can not broadcast message %s: %v", msg, err)
				}
			}
		}(ss)

	}
	wg.Wait()
	log.Printf("Finished broadcasting to %d clients: message %s", len(clients), msg)
}

// scope

func (s server) Broadcast(in *api.BroadcastSubscription, bsv api.ChatService_BroadcastServer) error {
	finished := make(chan bool)
	err := s.addBroadcastChannelToClient(in.Receiver, bsv, finished)
	if err != nil {
		return err
	}
	for {
		select {
		case <-finished:
			log.Printf("Closing stream for client %s", in.Receiver)
			return nil
		case <-bsv.Context().Done():
			log.Printf("Client %s has disconnected", in.Receiver)
			return nil
		}
	}
}

func (s *server) addBroadcastChannelToClient(id string, cs api.ChatService_BroadcastServer, fin chan bool) error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	_, ok := s.Clients[id]
	if !ok {
		return errors.New("Client has not joined yet")
	}
	s.Clients[id] = BroadcastSubscription{
		stream:   cs,
		finished: fin,
	}
	log.Printf("[%s: %d] Added a subscription to Broadcast for client %s", s.getName(), s.GetLamport(), id)
	return nil
}
