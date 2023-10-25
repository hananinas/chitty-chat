package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hananinas/chitty-chat/api"
	"github.com/hananinas/chitty-chat/internal/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	addr = "localhost:4321"
	name = "chit-chat"
)

var (
	addrFlag = flag.String("address", addr, "Enter your address, that you want to use on your chat-client")
	nameFlag = flag.String("name", fmt.Sprintf("%s-%d", name, time.Now().Unix()), "Enter the username you want to use on your chat-client")
	Lamport  = chat.LamportClock{Node: *nameFlag}
)

func main() {
	flag.Parse()

	log.Printf("Starting chat-client with name: %s and address: %s", *nameFlag, *addrFlag)
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addrFlag, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := api.NewChatServiceClient(conn)
	fmt.Printf("c: %v\n", c)

}

func join(client api.ChatServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := client.Join(ctx, &api.JoinRequest{})
	if err != nil {
		log.Fatalf("could not message: %v", err)
	}
}
