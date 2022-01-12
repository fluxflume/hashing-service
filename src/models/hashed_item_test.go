package models

import (
	"testing"
)

func TestHashedItem_GetHash(t *testing.T) {
	hashedItem := NewHashedItem(1, "123456789_abcdefghi_œ∑´´†¥_!@#$%^&")
	expectedValue := "N2U1ZjI5YzcxNTVkYTFmOWJiNTZhNjY3ZjcyNjRkZWViM2M2NmFkNzQxOWE5MDA3ZDBhZjcyYjgzMWM4YWVjZQ=="
	if expectedValue != hashedItem.GetHash() {
		t.Errorf("Hased value is incorrect")
	}
}

func TestHashedItem_GetId(t *testing.T) {
	hashedItem := NewHashedItem(100, "123456789_abcdefghi_œ∑´´†¥_!@#$%^&")
	expectedValue := uint64(100)
	if expectedValue != hashedItem.GetId() {
		t.Errorf("Hased value is incorrect")
	}
}
