package proto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/google/uuid"
)

type Command uint8
type QueueName [16]byte

const (
	ENQUEUE Command = iota
	ENQUEUE_ACK
	PEEK
	PEEK_ACK
	ACCEPT
	ACCEPT_ACK
	RELEASE
	RELEASE_ACK
	SIZE
	SIZE_ACK
)

const (
	//	MAX_PACKET_LENGTH = 4096
	MIN_PACKET_LENGTH = 17
)

type Queuic struct {
	Command   Command
	QueueName QueueName
	QueuicItem
}

type QueuicItem struct {
	Id   uuid.UUID
	Item []byte
}

func (q *QueueName) String() string {
	b := q[:]
	b = bytes.Trim(b, "\x00")
	return string(b[:])
}

func (q *QueueName) ParseFromString(s string) error {
	b := []byte(s)
	b = bytes.Trim(b, "\x00")
	n := copy(q[:], b[:16])
	if n < 16 {
		return fmt.Errorf("string %v to long", s)
	}
	return nil
}

func Encode(q *Queuic) ([]byte, error) {
	length := MIN_PACKET_LENGTH
	if q.QueuicItem.Item != nil {
		length += len(q.QueuicItem.Id)
		length += len(q.QueuicItem.Item)
	}
	b := make([]byte, length)
	b[0] = byte(q.Command)
	copy(b[1:17], q.QueueName[:])
	if q.QueuicItem.Item == nil {
		return b, nil
	} else {
		copy(b[17:33], q.QueuicItem.Id[:])
		copy(b[33:], q.QueuicItem.Item[:])
		return b, nil
	}
}

func Decode(data []byte) (*Queuic, error) {
	if len(data) < MIN_PACKET_LENGTH {
		return nil, fmt.Errorf("packet is too short")
	}
	/* 	if len(data) > MAX_PACKET_LENGTH {
		return nil, fmt.Errorf("packet is too long")
	} */
	var q Queuic
	q.Command = Command(data[0])
	copy(q.QueueName[:], data[1:17])
	if len(data) > MIN_PACKET_LENGTH {
		itemId, err := uuid.FromBytes(data[17:33])
		if err != nil {
			return nil, fmt.Errorf("failed to decode uuid: %v", err)
		}
		q.QueuicItem.Id = itemId
		q.QueuicItem.Item = data[33:]
	}
	return &q, nil
}

func Encrypt(key []byte, message []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes long")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gcm: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, message, nil)
	return ciphertext, nil
}

func Decrypt(key []byte, encryptedMessage []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes long")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gcm: %v", err)
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := encryptedMessage[:nonceSize], encryptedMessage[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %v", err)
	}
	return plaintext, nil
}
