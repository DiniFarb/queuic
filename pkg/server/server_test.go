package server_test

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/dinifarb/queuic/pkg/proto"
	"github.com/dinifarb/queuic/pkg/server"
	"github.com/google/uuid"
)

func TestCreateServerAndEnqueue(t *testing.T) {
	fileName := "./data/test.queuic"
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		os.Remove(fileName)
	}
	svr := server.NewQueuicServer("test")
	go func() {
		if err := svr.Serve(); err != nil {
			t.Errorf("server error: %v", err)
		}
	}()
	name := proto.QueueName{}
	copy(name[:], []byte("test"))
	if err := svr.CreateQueue(name); err != nil {
		t.Errorf("%v", err)
	}
	queueName := proto.QueueName{}
	copy(queueName[:], []byte("test"))
	req := proto.Queuic{
		Command:   proto.ENQUEUE,
		QueueName: queueName,
		QueuicItem: proto.QueuicItem{
			Id:   uuid.New(),
			Item: []byte("test message"),
		},
	}
	reqBytes, err := proto.Encode(&req)
	if err != nil {
		t.Errorf("failed to encode request: %v", err)
	}
	resp, err := sendUdpMessage(reqBytes)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	respQueuic, err := proto.Decode(resp)
	if err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if respQueuic.Command != proto.ENQUEUE_ACK {
		t.Errorf("unexpected response command: %v", respQueuic.Command)
	}
	peek := proto.Queuic{
		Command:   proto.PEEK,
		QueueName: queueName,
	}
	peekBytes, err := proto.Encode(&peek)
	if err != nil {
		t.Errorf("failed to encode request: %v", err)
	}
	resp, err = sendUdpMessage(peekBytes)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	respQueuic, err = proto.Decode(resp)
	if err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if respQueuic.Command != proto.PEEK_ACK {
		t.Errorf("unexpected response command: %v", respQueuic.Command)
	}
	accept := proto.Queuic{
		Command:   proto.ACCEPT,
		QueueName: queueName,
		QueuicItem: proto.QueuicItem{
			Id: respQueuic.QueuicItem.Id,
		},
	}
	acceptBytes, err := proto.Encode(&accept)
	if err != nil {
		t.Errorf("failed to encode request: %v", err)
	}
	resp, err = sendUdpMessage(acceptBytes)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	respQueuic, err = proto.Decode(resp)
	if err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if respQueuic.Command != proto.ACCEPT_ACK {
		t.Errorf("unexpected response command: %v", respQueuic.Command)
	}
}

func sendUdpMessage(send []byte) ([]byte, error) {
	key := sha256.Sum256([]byte("test"))
	fmt.Printf("client key: %x\n", key)
	encrypted, err := proto.Encrypt(key[:], send)
	if err != nil {
		return nil, err
	}
	s, err := net.ResolveUDPAddr("udp4", "localhost:9523")
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	_, err = c.Write(encrypted)
	if err != nil {
		return nil, err
	}
	fmt.Println("sent UDP Message, len: ", len(encrypted))
	buffer := make([]byte, 1024)
	n, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		fmt.Println("received UDP Message, len: ", len(buffer))
		decrypted, err := proto.Decrypt(key[:], buffer[:n])
		if err != nil {
			return nil, err
		}
		return decrypted, nil
	} else {
		return nil, fmt.Errorf("udp failed")
	}
}
