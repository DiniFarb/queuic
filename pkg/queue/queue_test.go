package queue_test

import (
	"os"
	"sync"
	"testing"

	"github.com/dinifarb/mlog"
	"github.com/dinifarb/queuic/pkg/proto"
	"github.com/dinifarb/queuic/pkg/queue"
	"github.com/google/uuid"
)

func TestQueueEnqueuePeekAccept(t *testing.T) {
	fileName := "./data/epa.queuic"
	mlog.SetLevel(mlog.Linfo)
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		os.Remove(fileName)
	}
	name := proto.QueueName{}
	copy(name[:], []byte("epa"))
	q, err := queue.NewQueue(name)
	if err != nil {
		mlog.Error("%v", err)
		os.Exit(1)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 0; i < 100; i++ {
			q.Enqueue(proto.QueuicItem{Id: uuid.New(), Item: []byte("test1")})
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 100; i++ {
			q.Enqueue(proto.QueuicItem{Id: uuid.New(), Item: []byte("test2")})
		}
		wg.Done()
	}()
	wg.Wait()
	if q.Size() != 200 {
		t.Errorf("Expected size 200, got %d", q.Size())
	}
	wg.Add(2)
	go func() {
		for i := 0; i < 100; i++ {
			i, err := q.Peek()
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			err = q.Accept(i.Id)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 100; i++ {
			i, err := q.Peek()
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			err = q.Accept(i.Id)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		}
		wg.Done()
	}()
	wg.Wait()
	if q.Size() != 0 {
		t.Errorf("Expected size 0, got %d", q.Size())
	}
	mlog.SetLevel(mlog.Ltrace)
}
