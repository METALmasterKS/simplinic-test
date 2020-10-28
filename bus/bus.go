package bus

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
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

	logger zerolog.Logger

	capacity int

	stack chan Message

	subscribers map[string]chan<- Message
	m           sync.RWMutex
}

// NewBus
func NewBus(ctx context.Context, logger zerolog.Logger, capacity int) *bus {
	var bus = &bus{
		stopCh:      make(chan bool),
		logger:      logger,
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
	if _, found := b.subscribers[name]; !found {
		return errors.New("subscriber not exists")
	}
	close(b.subscribers[name])

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
	b.logger.Info().Msg("start")
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
			b.Close()
			if len(b.stack) == 0 {
				b.logger.Info().Int("stack", len(b.stack)).Msg("stop")
				return
			}
			b.logger.Info().Int("stack", len(b.stack)).Msg("waiting for stop...")
		}
	}
}
