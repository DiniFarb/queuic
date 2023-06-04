package server

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dinifarb/mlog"
	"github.com/dinifarb/queuic/pkg/proto"
	"github.com/dinifarb/queuic/pkg/queue"
)

const (
	NETWORK_TYPE      = "udp"
	MAX_PACKET_LENGTH = 4096
	DEFAULT_PORT      = 9523
)

type QueuicServer struct {
	Port       int
	Key        [32]byte
	shutdown   chan bool
	queueStore QueueStore
}

type QueueStore struct {
	sync.RWMutex
	queues map[proto.QueueName]*queue.Queue
}

func NewQueuicServer(key string) *QueuicServer {
	q := make(map[proto.QueueName]*queue.Queue)
	k := sha256.Sum256([]byte(key))
	mlog.Debug("server key: %x", k)
	return &QueuicServer{
		Key:        k,
		queueStore: QueueStore{queues: q},
	}
}

func (s *QueuicServer) CreateQueue(name proto.QueueName) error {
	s.queueStore.Lock()
	defer s.queueStore.Unlock()
	for _, q := range s.queueStore.queues {
		if q.Name == name {
			return fmt.Errorf("queue %s already exists", name)
		}
	}
	q, err := queue.NewQueue(name)
	if err != nil {
		return fmt.Errorf("failed to create queue: %v", err)
	}
	s.queueStore.queues[name] = q
	mlog.Info("created queue: %s", name)
	return nil
}

//TODO DeleteQueue, LoadQueuesFromDisk

func (s *QueuicServer) Serve() error {
	s.shutdown = make(chan bool)
	if s.Key == [32]byte{} {
		return fmt.Errorf("server has no key source")
	}
	if s.Port == 0 {
		s.Port = DEFAULT_PORT
	}
	mlog.Info("receive on port: %d", s.Port)
	conn, err := net.ListenUDP(NETWORK_TYPE, &net.UDPAddr{Port: s.Port})
	if err != nil {
		return fmt.Errorf("listen to UDP failed with: %v", err)
	}
	defer conn.Close()
	for {
		var buff = make([]byte, MAX_PACKET_LENGTH)
		n, remoteAddr, err := conn.ReadFromUDP(buff[:])
		if err != nil {
			mlog.Error("error reading from connection: %v", err)
			continue
		}
		// breake if we got shutdown signal
		select {
		case <-s.shutdown:
			mlog.Info("shutdown signal received, shutting down")
			return nil
		default:
			go func(buff []byte, remoteAddr *net.UDPAddr) {
				mlog.Debug("received message from %s", remoteAddr)
				decryptedMessage, err := proto.Decrypt(s.Key[:], buff)
				if err != nil {
					mlog.Error("error decrypting message: %v", err)
					return
				}
				resp, err := s.HandleQueuicRequest(decryptedMessage)
				if err != nil {
					mlog.Error("error handling request: %v", err)
					return
				}
				encryptedMessage, err := proto.Encrypt(s.Key[:], resp)
				if err != nil {
					mlog.Error("error encrypting message: %v", err)
					return
				}
				mlog.Debug("write message back to %s", remoteAddr)
				_, err = conn.WriteToUDP(encryptedMessage, remoteAddr)
				if err != nil {
					mlog.Error("error writing to connection: %v", err)
				}

			}(append([]byte(nil), buff[:n]...), remoteAddr)
		}
	}
}

func (s *QueuicServer) Shutdown() {
	// TODO Implement graceful shutdown
	// wait for all onging requests to finish
	s.shutdown <- true
	time.Sleep(5 * time.Second)
	os.Exit(0)
}
