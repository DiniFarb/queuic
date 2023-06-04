package proto_test

import (
	"crypto/sha256"
	"testing"

	"github.com/dinifarb/queuic/pkg/proto"
	"github.com/google/uuid"
)

func TestEnDecode(t *testing.T) {
	queueName := proto.QueueName{}
	copy(queueName[:], []byte("test"))
	q := proto.Queuic{
		Command:   proto.ENQUEUE,
		QueueName: queueName,
		QueuicItem: proto.QueuicItem{
			Id:   uuid.New(),
			Item: []byte("test message"),
		},
	}
	b, err := proto.Encode(&q)
	if err != nil {
		t.Errorf("failed to encode request: %v", err)
	}
	q2, err := proto.Decode(b)
	if err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if q2.Command != proto.ENQUEUE {
		t.Errorf("unexpected response command: %v", q2.Command)
	}
	if q2.QueueName != queueName {
		t.Errorf("unexpected queue name: %v", q2.QueueName)
	}
	if string(q2.QueuicItem.Item) != "test message" {
		t.Errorf("unexpected value: %v", q2.QueuicItem.Item)
	}
}

func TestEnDecodeWithCrypto(t *testing.T) {
	queueName := proto.QueueName{}
	copy(queueName[:], []byte("test"))
	q := proto.Queuic{
		Command:   proto.ENQUEUE,
		QueueName: queueName,
		QueuicItem: proto.QueuicItem{
			Id:   uuid.New(),
			Item: []byte("test message"),
		},
	}
	b, err := proto.Encode(&q)
	if err != nil {
		t.Errorf("failed to encode request: %v", err)
	}
	enk := sha256.Sum256([]byte("test"))
	encrypted, err := proto.Encrypt(enk[:], b)
	if err != nil {
		t.Errorf("failed to encrypt request: %v", err)
	}
	dek := sha256.Sum256([]byte("test"))
	decrypted, err := proto.Decrypt(dek[:], encrypted)
	if err != nil {
		t.Errorf("failed to decrypt request: %v", err)
	}
	q2, err := proto.Decode(decrypted)
	if err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if q2.Command != proto.ENQUEUE {
		t.Errorf("unexpected response command: %v", q2.Command)
	}
	if q2.QueueName != queueName {
		t.Errorf("unexpected queue name: %v", q2.QueueName)
	}
	if string(q2.QueuicItem.Item) != "test message" {
		t.Errorf("unexpected value: %v", q2.QueuicItem.Item)
	}
}
