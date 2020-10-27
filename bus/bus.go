package bus

import (
	"context"
	"errors"
	"sync"
)

//Broker
type Broker interface {
	Publish(ctx context.Context, msg Message) error
	Subscribe(name string) (<-chan Message, error)
	Unsubscribe(name string) error
	Close()
}

// bus
type bus struct {
	stopCh chan bool

	capacity int

	stack chan Message

	subscribers map[string]chan<- Message
	m           sync.RWMutex
}

// NewBus
func NewBus(ctx context.Context, capacity int) *bus {
	var bus = &bus{
		stopCh:      make(chan bool),
		subscribers: make(map[string]chan<- Message),
		stack:       make(chan Message, capacity),
		capacity:    capacity,
	}

	go bus.process(ctx)

	return bus
}

func (b *bus) Close() {
	select {
	case <-b.stopCh:
		return
	default:
		close(b.stopCh)
	}
}

func (b *bus) Publish(ctx context.Context, msg Message) error {
	select {
	case <-b.stopCh:
		return errors.New("broker closed")
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	b.stack <- msg

	return nil
}

func (b *bus) Subscribe(name string) (<-chan Message, error) {
	select {
	case <-b.stopCh:
		return nil, errors.New("broker closed")
	default:
	}

	ch := make(chan Message, b.capacity)
	b.m.Lock()
	b.subscribers[name] = ch
	b.m.Unlock()
	return ch, nil
}

func (b *bus) Unsubscribe(name string) error {
	select {
	case <-b.stopCh:
		return errors.New("broker closed")
	default:
	}

	b.m.Lock()
	delete(b.subscribers, name)
	b.m.Unlock()
	return nil
}

func (b *bus) broadcast(msg Message) {
	var subscribers map[string]chan<- Message

	b.m.RLock()
	subscribers = b.subscribers
	b.m.RUnlock()

	for name := range subscribers {
		select {
		case subscribers[name] <- msg:
		case <-b.stopCh:
			return
		}
	}

}

func (b *bus) process(ctx context.Context) {
	semaphore := make(chan struct{}, 10)
	for {
		select {
		case msg := <-b.stack:
			semaphore <- struct{}{}
			go func(msg Message) {
				defer func() {
					<-semaphore
				}()
				b.broadcast(msg)
			}(msg)
		case <-ctx.Done():
			return
		}
	}
}
