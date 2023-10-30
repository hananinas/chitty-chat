package client

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hananinas/chitty-chat/api"
	"github.com/hananinas/chitty-chat/internal/chat"
	"github.com/manifoldco/promptui"
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
	lamport  = chat.LamportClock{Node: *nameFlag}
)

func StartClient(nameInput string, addrInput string) {

	if nameInput != "" || addrInput != "" {
		*nameFlag = nameInput
		*addrFlag = addr
	}

	log.Printf("%s", *addrFlag)

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addrFlag, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := api.NewChatServiceClient(conn)
	join(c)
	go broadcastListener(c)

	activeChat(c)
}

// client sends join request
func join(client api.ChatServiceClient) {

	log.Printf("Joining chat with -- Name: %s --- Time %d", *nameFlag, lamport.GetTimestamp())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lamport.Move()

	res, err := client.Join(ctx, &api.JoinRequest{
		NodeName: *nameFlag,
		Lamport:  &api.Lamport{Time: lamport.GetTimestamp(), NodeId: *nameFlag},
	})
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	lamport.CompOtherClock(res.Lamport.GetTime())

	log.Printf("Hey user [%s] you joined chat at Time [%d] with status %s", *nameFlag, lamport.GetTimestamp(), res.Status)

}

// leave chat
func leave(client api.ChatServiceClient) {
	log.Printf("Leaving chat with name: %s -- time %d", *nameFlag, lamport.GetTimestamp())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Leave(ctx, &api.LeaveRequest{
		SenderId: *nameFlag,
		Lamport:  &api.Lamport{Time: lamport.GetTimestamp(), NodeId: *nameFlag},
	})
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	if res.GetStatus() != api.Status_OK {
		log.Fatalf("could not leave: %v", res.GetStatus())
	}
	lamport.CompOtherClock(res.Lamport.GetTime())

	log.Printf("Left chat with name: %s", *nameFlag)

}

// now that user can join and leave the user should now be able to stay in the chat and chat and input messages

func activeChat(client api.ChatServiceClient) {
	active := true

	for active {
		prompt := promptui.Select{
			Label: "Select Action",
			Items: []string{"Leave", "chat"},
		}
		_, input, err := prompt.Run()
		if err != nil {
			log.Fatalf("could not get input: %v", err)
		}

		if input == "Leave" {
			active = false
			leave(client)
		}

		lamport.Move()

		if input == "chat" {
			if active {
				prompt := promptui.Prompt{
					Label: "input your message and send ",
					Validate: func(input string) error {
						if len(input) == 0 || len(input) > 100 {
							return fmt.Errorf("input must be at least 1 character or under 100 characters")
						}
						return nil
					},
				}
				input, err := prompt.Run()
				if err != nil {
					log.Fatalf("could not get input: %v", err)
				}

				send(client, input)
			}
		}

	}

}

// func listen for broadcast

// reacive the broadcast stream from the server
func broadcastListener(client api.ChatServiceClient) {
	log.Printf("listening for broadcast")
	ctx := context.Background()

	stream, err := client.Broadcast(ctx, &api.BroadcastSubscription{Receiver: *nameFlag})
	if err != nil {
		log.Fatalf("[Node %s]Could not subscribe to broadcast stream: %v", *nameFlag, err)
	}

	go handleMessages(stream)
}

func handleMessages(stream api.ChatService_BroadcastClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("error receiving message: %v", err)
			return
		}

		lamport.CompOtherClock(msg.Lamport.GetTime())
		log.Printf("Node [%s] at Time [%d] received broadcasted message --- %s ", *nameFlag, lamport.GetTimestamp(), msg.GetContent())
	}
}

// func send a chat message

func send(client api.ChatServiceClient, msg string) {
	lamport.Move()
	log.Printf("[Node %s: %d] Sending message >>> %s", *nameFlag, lamport.GetTimestamp(), msg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Send(ctx, &api.Message{
		Lamport: &api.Lamport{Time: lamport.GetTimestamp(), NodeId: *nameFlag},
		Content: msg,
	})
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	if res.GetStatus() != api.Status_OK {
		log.Fatalf("could not send message: %v", res.GetStatus())
	}

	lamport.CompOtherClock(res.Lamport.GetTime())

	log.Printf("Node [%s] at Time [%d] sent message --- %s ", *nameFlag, lamport.GetTimestamp(), msg)
}
