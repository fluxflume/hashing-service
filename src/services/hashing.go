// Contains definitions related to the hashing service component, as well as default implementations.
// This is the main component used by the http handler.

package services

import (
	"fmt"
	"hashing/src/models"
)

// Hashing value restrictions:
// 1. No sense in dealing empty strings.
// 2. Cap at something reasonable for the reasons described here: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#implement-proper-password-strength-controls
var minLength, maxLength = 1, 64

type HashingService interface {
	Create(value string) (models.HashedItem, error)
	Get(id uint64) (models.HashedItem, error)
}

type InputLengthError struct {
	length int
}

func (e *InputLengthError) Error() string {
	return fmt.Sprintf("Unsupported input length: %d. "+
		"Must be at least %d and no more than %d.", e.length, minLength, maxLength)
}

//

type DefaultHashingService struct {
	publisher   CreatedEventPublisher
	store       HashedItemStore
	currentId   uint64
	idGenerator IdGenerator
}

func (receiver *DefaultHashingService) Create(value string) (models.HashedItem, error) {
	length := len(value)
	if length < minLength || length > maxLength {
		return models.HashedItem{}, &InputLengthError{length: length}
	}
	id := receiver.idGenerator.Get()
	item := models.NewHashedItem(id, value)
	if err := receiver.publisher.Publish(item); err != nil {
		return models.HashedItem{}, err
	} else {
		return item, nil
	}
}

func (receiver *DefaultHashingService) Get(id uint64) (models.HashedItem, error) {
	return receiver.store.Get(id)
}

func CreateDefaultHashingService(publisher CreatedEventPublisher, store HashedItemStore, idGenerator IdGenerator) HashingService {
	return &DefaultHashingService{publisher: publisher, store: store, idGenerator: idGenerator}
}
