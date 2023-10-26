package chat

import "github.com/hananinas/chitty-chat/api"

type BroadcastSubscription struct {
	stream   api.ChatService_BroadcastServer
	finished chan<- bool
}
