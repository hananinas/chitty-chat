package chat

import (
	"log"
	"sync"
)

// LamportClock is a struct that represents a Lamport clock.
type LamportClock struct {
	mu    sync.Mutex
	value uint32
	Node  string
}

func (c *LamportClock) Move() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
	log.Printf("User %s Timestamp:  %d", c.Node, c.value)
}

func (c *LamportClock) GetTimestamp() uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func (c *LamportClock) CompOtherClock(timestamp uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = max(c.value, timestamp) + 1
	log.Printf("User %s Timestamp:  %d", c.Node, c.value)
}

func max(a, b uint32) uint32 {
	if a < b {
		return b
	}
	return a
}
