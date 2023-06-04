package server

import (
	"encoding/binary"
	"fmt"

	"github.com/dinifarb/mlog"
	"github.com/dinifarb/queuic/pkg/proto"
	"github.com/dinifarb/queuic/pkg/queue"
	"github.com/google/uuid"
)

func (s *QueuicServer) HandleQueuicRequest(b []byte) ([]byte, error) {
	req, err := proto.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request: %v", err)
	}
	queue, ok := s.queueStore.queues[req.QueueName]
	if !ok {
		//TODO: handle create queue on the fly
		return nil, fmt.Errorf("queue %s does not exists", req.QueueName)
	}
	switch req.Command {
	case proto.ENQUEUE:
		return handleEnqueue(queue, req)
	case proto.PEEK:
		return handlePeek(queue, req)
	case proto.ACCEPT:
		return handleAccept(queue, req)
	case proto.RELEASE:
		return handleRelease(queue, req)
	case proto.SIZE:
		return handleSize(queue, req)
	default:
		return nil, fmt.Errorf("unknown command: %v", req.Command)
	}
}

func handleEnqueue(current_queue *queue.Queue, q *proto.Queuic) ([]byte, error) {
	err := current_queue.Enqueue(q.QueuicItem)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue: %v", err)
	}
	mlog.Debug("enqueued item: %v", q.QueuicItem.Id)
	ack := proto.Queuic{
		Command:   proto.ENQUEUE_ACK,
		QueueName: q.QueueName,
	}
	return encodeResponse(&ack)
}

func handlePeek(current_queue *queue.Queue, q *proto.Queuic) ([]byte, error) {
	queueItem, err := current_queue.Peek()
	if err != nil {
		return nil, fmt.Errorf("failed to peek: %v", err)
	}
	mlog.Debug("peeked item: %v", queueItem.Id)
	ack := proto.Queuic{
		Command:    proto.PEEK_ACK,
		QueueName:  q.QueueName,
		QueuicItem: queueItem,
	}
	return encodeResponse(&ack)
}

func handleAccept(current_queue *queue.Queue, q *proto.Queuic) ([]byte, error) {
	err := current_queue.Accept(q.QueuicItem.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to accept: %v", err)
	}
	mlog.Debug("accepted item: %v", q.QueuicItem.Id)
	ack := proto.Queuic{
		Command:   proto.ACCEPT_ACK,
		QueueName: q.QueueName,
	}
	return encodeResponse(&ack)
}

func handleRelease(current_queue *queue.Queue, q *proto.Queuic) ([]byte, error) {
	err := current_queue.Release(q.QueuicItem.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to release: %v", err)
	}
	ack := proto.Queuic{
		Command:   proto.ACCEPT_ACK,
		QueueName: q.QueueName,
	}
	return encodeResponse(&ack)
}

// TODO: handle size of queue
func handleSize(current_queue *queue.Queue, q *proto.Queuic) ([]byte, error) {
	size := current_queue.Size()
	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(size))
	ack := proto.Queuic{
		Command:   proto.SIZE_ACK,
		QueueName: q.QueueName,
		QueuicItem: proto.QueuicItem{
			Id:   uuid.New(),
			Item: sizeBytes,
		},
	}
	return encodeResponse(&ack)
}

func encodeResponse(q *proto.Queuic) ([]byte, error) {
	b, err := proto.Encode(q)
	if err != nil {
		return nil, fmt.Errorf("failed to encode response: %v", err)
	}
	return b, nil
}
