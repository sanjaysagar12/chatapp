package services

import "github.com/turplespace/msgqueue/internal/models"

type QueueService struct {
	handler *models.WebSocketHandler
}

// NewQueueService creates a new QueueService instance
func NewQueueService(handler *models.WebSocketHandler) *QueueService {
	return &QueueService{
		handler: handler,
	}
}

func (qs *QueueService) EnQueue(topic string, message string) {
	qs.handler.Mu.Lock()
	defer qs.handler.Mu.Unlock()
	qs.handler.Queue[topic] = append(qs.handler.Queue[topic], message)
}

func (qs *QueueService) DeQueue(topic string) (string, bool) {
	qs.handler.Mu.Lock()
	defer qs.handler.Mu.Unlock()
	if len(qs.handler.Queue[topic]) == 0 {
		return "", false
	}
	msg := qs.handler.Queue[topic][0]
	qs.handler.Queue[topic] = qs.handler.Queue[topic][1:]
	return msg, true
}
