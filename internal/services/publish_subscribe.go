package services

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/turplespace/msgqueue/internal/models"
)

type PublichServics struct {
	handler *models.WebSocketHandler
}

func NewPublishService(handler *models.WebSocketHandler) *PublichServics {
	return &PublichServics{
		handler: handler,
	}
}
func (ps *PublichServics) AddSubscribers(topic string, conn *websocket.Conn) {
	ps.handler.Mu.Lock()
	defer ps.handler.Mu.Unlock()
	ps.handler.Subscribers[topic] = append(ps.handler.Subscribers[topic], conn)
}

func (ps *PublichServics) RemoveConnection(conn *websocket.Conn) {
	ps.handler.Mu.Lock()
	defer ps.handler.Mu.Unlock()
	for topic, connections := range ps.handler.Subscribers {
		var updatedConnections []*websocket.Conn
		for _, c := range connections {
			if c != conn {
				updatedConnections = append(updatedConnections, c)
			}
		}
		ps.handler.Subscribers[topic] = updatedConnections
	}
}

func (ps *PublichServics) SendMessageToSubscribers(message models.Message) {
	ps.handler.Mu.RLock()
	defer ps.handler.Mu.RUnlock()
	connections, exists := ps.handler.Subscribers[message.Topic]
	if !exists {
		return
	}
	for _, conn := range connections {
		err := conn.WriteJSON(message.Message)
		if err != nil {
			log.Println("Writing Error:", err)
			ps.RemoveConnection(conn)
		}
	}
}
