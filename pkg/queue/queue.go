package queue

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"sync"

	"github.com/dinifarb/mlog"
	"github.com/dinifarb/queuic/pkg/proto"
	"github.com/google/uuid"
)

const (
	path = "./data/%s.queuic"
)

type Queue struct {
	items   []proto.QueuicItem
	peeked  map[uuid.UUID]proto.QueuicItem
	mu      sync.Mutex
	store   store
	added   uint64
	removed uint64
	Name    proto.QueueName
}

type store struct {
	file os.File
	mu   sync.Mutex
}

func NewQueue(name proto.QueueName) (*Queue, error) {
	q := &Queue{
		Name: name,
	}
	q.items = make([]proto.QueuicItem, 0)
	q.peeked = make(map[uuid.UUID]proto.QueuicItem)
	fileName := fmt.Sprintf(path, name.String())
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		f, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create bin file: %w", err)
		}
		q.store.file = *f
	} else {
		f, err := os.OpenFile(fileName, os.O_RDWR, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open bin file: %w", err)
		}
		q.store.file = *f
		if err := q.loadFromDisk(); err != nil {
			return nil, fmt.Errorf("failed to load from disk: %w", err)
		}
	}
	return q, nil
}

func (q *Queue) Enqueue(item proto.QueuicItem) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
	q.added++
	return q.saveToDisk()
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items) + len(q.peeked)
}

func (q *Queue) Enqueued() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.added
}

func (q *Queue) Dequeued() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.removed
}

func (q *Queue) Peek() (proto.QueuicItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return proto.QueuicItem{}, fmt.Errorf("queue is empty")
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.peeked[item.Id] = item
	err := q.saveToDisk()
	if err != nil {
		return proto.QueuicItem{}, err
	}
	return item, nil
}

func (q *Queue) Release(id uuid.UUID) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	//copy peeked to to 0 on items
	q.items = append([]proto.QueuicItem{q.peeked[id]}, q.items...)
	delete(q.peeked, id)
	return q.saveToDisk()
}

func (q *Queue) Accept(id uuid.UUID) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.peeked, id)
	q.removed++
	return q.saveToDisk()
}

func (q *Queue) Delete() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.store.file.Close()
	fileName := fmt.Sprintf(path, q.Name.String())
	if err := os.Remove(fileName); err != nil {
		return fmt.Errorf("failed to remove queue file: %w", err)
	}
	return nil
}

func (q *Queue) saveToDisk() error {
	q.store.mu.Lock()
	defer q.store.mu.Unlock()
	if len(q.items) == 0 && len(q.peeked) == 0 {
		return q.replaceFile()
	}
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(q.items); err != nil {
		return fmt.Errorf("gob error: %w", err)
	}
	if _, err := q.store.file.Write(buff.Bytes()); err != nil {
		return fmt.Errorf("failed to write bytes to disk: %w", err)
	}
	mlog.Debug("saved to disk - items %d, peeked %d", len(q.items), len(q.peeked))
	return nil
}

func (q *Queue) loadFromDisk() error {
	q.store.mu.Lock()
	defer q.store.mu.Unlock()
	dec := gob.NewDecoder(&q.store.file)
	if err := dec.Decode(&q.items); err != nil {
		if err.Error() == "EOF" {
			//EMPTY FILE
			return nil
		}
		return fmt.Errorf("failed to read bytes from disk: %w", err)
	}
	return nil
}

func (q *Queue) replaceFile() error {
	q.store.file.Close()
	fileName := fmt.Sprintf(path, q.Name.String())
	if err := os.Remove(fileName); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	_, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	return nil
}
