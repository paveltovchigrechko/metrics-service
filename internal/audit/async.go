package audit

import (
	"errors"
	"io"
	"log"
	"sync"
)

// ErrQueueFull is returned when the async queue has no room and the event is dropped.
var ErrQueueFull = errors.New("audit queue is full, event dropped")

// AsyncObserver decorates an Observer: WriteEvent only enqueues the event and returns immediately,
// a single worker goroutine delivers events to the wrapped observer in order.
// Each wrapped observer has its own queue, so a slow receiver does not delay the others.
type AsyncObserver struct {
	next   Observer
	events chan *Event
	done   chan struct{}

	mu     sync.RWMutex // Guards closed and sending to events against Close.
	closed bool

	closeNext sync.Once
	closeErr  error
}

func NewAsyncObserver(next Observer, buffer int) *AsyncObserver {
	a := &AsyncObserver{
		next:   next,
		events: make(chan *Event, buffer),
		done:   make(chan struct{}),
	}

	go a.run()

	return a
}

func (a *AsyncObserver) run() {
	defer close(a.done)

	for e := range a.events {
		if err := a.next.WriteEvent(e); err != nil {
			log.Printf("audit: %v", err)
		}
	}
}

// WriteEvent never blocks. If the queue is full, the event is dropped and ErrQueueFull is returned.
func (a *AsyncObserver) WriteEvent(e *Event) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.closed {
		return errors.New("audit observer is closed")
	}

	select {
	case a.events <- e:
		return nil
	default:
		return ErrQueueFull
	}
}

// Close stops accepting events, waits until the queued ones are delivered and
// then closes the wrapped observer if it implements io.Closer.
// It is safe to call Close more than once; later calls return the first result.
func (a *AsyncObserver) Close() error {
	a.mu.Lock()
	if !a.closed {
		a.closed = true
		close(a.events)
	}
	a.mu.Unlock()

	<-a.done

	a.closeNext.Do(func() {
		if c, ok := a.next.(io.Closer); ok {
			a.closeErr = c.Close()
		}
	})

	return a.closeErr
}
