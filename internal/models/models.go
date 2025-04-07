package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Subscribe        bool   `json:"subscribe"`
	Role             string `json:"role"`
	Message          string `json:"message"`
	Topic            string `json:"topic"`
	TransmissionMode string `json:"transmission_mode"`
}

type WebSocketHandler struct {
	Upgrader    websocket.Upgrader
	Subscribers map[string][]*websocket.Conn
	Mu          sync.RWMutex
	Queue       map[string][]string
}
