package main

import (
	"flag"
	"fmt"
	"time"

	"github.itu.dk/agpr/chitty-chat/internal/chat"
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
}
