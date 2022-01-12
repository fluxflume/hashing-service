package models

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type HashedItem struct {
	id   uint64
	hash string
}

func (hashedItem HashedItem) GetId() uint64 {
	return hashedItem.id
}

func (hashedItem HashedItem) GetHash() string {
	return hashedItem.hash
}

func NewHashedItem(id uint64, value string) HashedItem {
	return HashedItem{
		id:   id,
		hash: hash(value),
	}
}

func encrypt(input string) string {
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)
}

func encode(input string) string {
	data := []byte(input)
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded
}

func hash(value string) string {
	return encode(encrypt(value))
}
