package services

import (
	"fmt"
	"hashing/src/models"
	"sync"
)

type HashedItemStore interface {
	Put(hashedEntity models.HashedItem) (models.HashedItem, error)
	Get(id uint64) (models.HashedItem, error)
}

//

type LocalHashedItemStore struct {
	items     map[uint64]models.HashedItem
	itemsLock *sync.RWMutex
}

func (localStore *LocalHashedItemStore) Get(id uint64) (models.HashedItem, error) {
	localStore.itemsLock.RLock()
	defer localStore.itemsLock.RUnlock()
	if item, ok := localStore.items[id]; ok {
		return item, nil
	} else {
		return models.HashedItem{}, fmt.Errorf("item with id %d does not exist", id)
	}
}

func (localStore *LocalHashedItemStore) Put(item models.HashedItem) (models.HashedItem, error) {
	localStore.itemsLock.Lock()
	defer localStore.itemsLock.Unlock()
	if _, exists := localStore.items[item.GetId()]; exists {
		return models.HashedItem{}, fmt.Errorf("item with id %d already exists", item.GetId())
	} else {
		localStore.items[item.GetId()] = item
		return item, nil
	}
}

func CreateLocalStore() HashedItemStore {
	return &LocalHashedItemStore{items: map[uint64]models.HashedItem{}, itemsLock: &sync.RWMutex{}}
}
