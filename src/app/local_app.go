package main

import (
	"hashing/src/http"
	"hashing/src/models"
	"hashing/src/services"
	"log"
	"sync"
	"time"
)

func StartLocalApp(port int, maxPendingItems int, numWorkers int, hashCompletionDelay time.Duration) {
	// The channel used for buffering incoming hash requests
	hashedItemChan := make(chan models.HashedItem, maxPendingItems)

	// A WaitGroup used for delaying shutdown until all existing items have been processed.
	pendingItemsWaitGroup := &sync.WaitGroup{}

	// A "local" publisher, which publishes hashed items to a channel.
	publisher := services.CreateLocalPublisher(pendingItemsWaitGroup, hashedItemChan)

	// Instance of the local store, which persists hashed items in-memory
	store := services.CreateLocalStore()

	// A "local" subscriber", which reads from the hashed item channel, waits, then saves into the "store"
	subscriber := services.CreateLocalSubscriber(pendingItemsWaitGroup, hashedItemChan, store, hashCompletionDelay)

	// Initialize the workers to begin pulling from the channel
	for i := 0; i < numWorkers; i++ {
		subscriber.Subscribe()
	}

	idGenerator := services.NewLocalIdGenerator()

	// An implementation of the hashing service component
	hashingService := services.CreateDefaultHashingService(publisher, store, idGenerator)

	// A simple, thread-safe stats component which tracks the number of total requests to POST /hash
	// and provides an aggregated average request/response time.
	statsService := services.CreateLocalStatsService()

	// Go!
	http.StartAndWaitUntilShutdown(port, hashingService, statsService)

	// The server has been shutdown... which means no new requests will be processed, but existing requests can complete...
	pendingItemsWaitGroup.Wait()
	log.Println("Completed")
}
