package main

import (
	"log"

	"github.com/hananinas/chitty-chat/internal/client"
	"github.com/hananinas/chitty-chat/internal/server"
	"github.com/manifoldco/promptui"
)

func main() {
	// add a promptui wich asks the user if he wants to start a server or a client
	// if server is chosen, start a server
	// if client is chosen, start a client
	prompt := promptui.Select{
		Label: "Select Action",
		Items: []string{"Start a New Server", "Create a New Client"},
	}
	_, input, err := prompt.Run()
	if err != nil {
		log.Fatalf("could not get input: %v", err)

	}

	if input == "Start a New Server" {
		// ask for a name for the server

		server.StartServer()
	}

	if input == "Create a New Client" {
		// ask for a name for the client and address flag else use default settings
		promptName := promptui.Prompt{
			Label: "Give a name for the client or leave empty for default settings",
		}

		name, err := promptName.Run()
		if err != nil {
			log.Fatalf("could not get input: %v", err)
		}

		if name == "" || len(name) < 50 {

			promptAddr := promptui.Prompt{
				Label: "Give a address for the client to run on or leave empty for default settings",
			}

			addr, err := promptAddr.Run()
			if err != nil {
				log.Fatalf("could not get input: %v", err)

			}
			client.StartClient(name, addr)

		}

	}
}
