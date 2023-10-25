package server

import (
	pb "github.com/hananinas/chitty-chat/api"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedChatServiceServer
}
