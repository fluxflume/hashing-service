package services

import (
	"fmt"
	"hashing/src/models"
	"log"
	"sync"
	"time"
)

type CreatedEventPublisher interface {
	Publish(hashedItem models.HashedItem) error
}

type CreatedEventSubscriber interface {
	Subscribe()
}

type PublishingError struct {
}

func (t *PublishingError) Error() string {
	return fmt.Sprintf("channel is closed or too many pending items; try again later")
}

//

type LocalPublisher struct {
	wg *sync.WaitGroup
	ch chan<- models.HashedItem
}

func (localPublisher LocalPublisher) Publish(hashedItem models.HashedItem) error {
	select {
	case localPublisher.ch <- hashedItem: // Send to the channel, unless it is full
		localPublisher.wg.Add(1)
		log.Printf("Publisher published item.id=%d", hashedItem.GetId())
		return nil
	default:
		log.Printf("Publisher failed to publish item.id=%d", hashedItem.GetId())
		return &PublishingError{}
	}
}

func CreateLocalPublisher(wg *sync.WaitGroup, ch chan<- models.HashedItem) CreatedEventPublisher {
	return LocalPublisher{wg: wg, ch: ch}
}

//

type LocalSubscriber struct {
	wg    *sync.WaitGroup
	ch    <-chan models.HashedItem
	store HashedItemStore
	delay time.Duration
}

func (localSubscriber *LocalSubscriber) Subscribe() {
	go func() {
		for {
			v := <-localSubscriber.ch
			t := time.Now()
			log.Printf("Subscriber received item.id=%d", v.GetId())
			time.Sleep(localSubscriber.delay)
			if _, err := localSubscriber.store.Put(v); err != nil {
				log.Printf("Subscriber failed to store item.id=%d in %d ms", v.GetId(), time.Since(t).Milliseconds())
			} else {
				log.Printf("Subscriber succesfully stored item.id=%d in %d ms", v.GetId(), time.Since(t).Milliseconds())
			}
			// Note: If this impl was using a messaging system, this subscriber could be on a different service instance
			// or running in a completely different service.
			// This means that the acknowledgement of the store.Put() above would need to be published as an event,
			// and then later consumed by the hashing service, allowing the WaitGroup.Done() call below to happen.
			// Additionally, since these WaitGroup's are stateful and unique to each hashing service instance,
			// a correlation id would be needed so that the correct service instance consumes/acknowledges the event,
			localSubscriber.wg.Done()
			// Note: This is the point at which a client of a messaging system would normally ack/nack the message.
			// Note: Since this is a mock pubsub system, there's no acking - if the store.Put call above fails, the message will be lost.
		}
	}()
}

func CreateLocalSubscriber(wg *sync.WaitGroup, ch <-chan models.HashedItem, store HashedItemStore, delay time.Duration) CreatedEventSubscriber {
	return &LocalSubscriber{wg: wg, ch: ch, store: store, delay: delay}
}
