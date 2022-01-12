package services

import (
	"fmt"
	"hashing/src/models"
	"sync/atomic"
	"testing"
)

// A stub for the CreatedEventPublisher
type CreatedEventPublisherMock struct {
	_publish func(hashedItem models.HashedItem) error
}

func (m *CreatedEventPublisherMock) Publish(hashedItem models.HashedItem) error {
	if m._publish != nil {
		return m._publish(hashedItem)
	}
	return nil
}

type HashedItemStoreMock struct {
	_put func(hashedEntity models.HashedItem) (models.HashedItem, error)
	_get func(id uint64) (models.HashedItem, error)
}

func (m *HashedItemStoreMock) Put(hashedEntity models.HashedItem) (models.HashedItem, error) {
	if m._put != nil {
		return m._put(hashedEntity)
	}
	return models.HashedItem{}, nil
}

func (m *HashedItemStoreMock) Get(id uint64) (models.HashedItem, error) {
	if m._get != nil {
		return m._get(id)
	}
	return models.HashedItem{}, nil
}

func NewCreatedEventPublisherMock(handler func(hashedItem models.HashedItem) error) CreatedEventPublisher {
	return &CreatedEventPublisherMock{
		_publish: handler,
	}
}

func NewHashedItemStoreMock(get func(id uint64) (models.HashedItem, error), put func(hashedEntity models.HashedItem) (models.HashedItem, error)) HashedItemStore {
	return &HashedItemStoreMock{_get: get, _put: put}
}

func TestStoreGetInteraction(t *testing.T) {
	publisher := NewCreatedEventPublisherMock(func(hashedItem models.HashedItem) error {
		return nil
	})
	item := models.NewHashedItem(1, "test")
	store := NewHashedItemStoreMock(func(id uint64) (models.HashedItem, error) {
		return item, nil
	}, nil)
	idGenerator := NewLocalIdGenerator()
	hashingService := CreateDefaultHashingService(publisher, store, idGenerator)
	result, err := hashingService.Get(1)
	if err != nil {
		t.Error(err)
	}
	if result.GetId() != item.GetId() {
		t.Errorf("Expected item with id=%d but got %d", item.GetId(), result.GetId())
	}
	if result.GetHash() != item.GetHash() {
		t.Errorf("Expected item with hash %s but got %s", item.GetHash(), result.GetHash())
	}
}

func TestStoreGetErrorInteraction(t *testing.T) {
	publisher := NewCreatedEventPublisherMock(func(hashedItem models.HashedItem) error {
		return nil
	})
	e := fmt.Errorf("Boom")
	store := NewHashedItemStoreMock(func(id uint64) (models.HashedItem, error) {
		return models.HashedItem{}, e
	}, nil)
	idGenerator := NewLocalIdGenerator()
	hashingService := CreateDefaultHashingService(publisher, store, idGenerator)
	_, err := hashingService.Get(1)
	if err == nil {
		t.Error("Expected error but received nil")
	}
	if e.Error() != err.Error() {
		t.Errorf("Expected error with message %s but got %s", e.Error(), err.Error())
	}
}

func TestPublisherPublishInteraction(t *testing.T) {
	received := atomic.Value{}
	item := models.NewHashedItem(0, "test")
	publisher := NewCreatedEventPublisherMock(func(hashedItem models.HashedItem) error {
		received.Store(hashedItem)
		return nil
	})
	store := NewHashedItemStoreMock(nil, nil)
	idGenerator := NewLocalIdGenerator()
	hashingService := CreateDefaultHashingService(publisher, store, idGenerator)
	result, err := hashingService.Create("test")
	if err != nil {
		t.Error(err)
	}
	if result.GetId() != item.GetId() {
		t.Errorf("Expected item with id=%d but got %d", item.GetId(), result.GetId())
	}
	if result.GetHash() != item.GetHash() {
		t.Errorf("Expected item with hash %s but got %s", item.GetHash(), result.GetHash())
	}
	publishedValue := received.Load().(models.HashedItem)
	if publishedValue.GetId() != item.GetId() {
		t.Errorf("Expected published item with id=%d but got %d", item.GetId(), publishedValue.GetId())
	}
	if publishedValue.GetHash() != item.GetHash() {
		t.Errorf("Expected published item with hash %s but got %s", item.GetHash(), publishedValue.GetHash())
	}

}
