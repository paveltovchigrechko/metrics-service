package audit

import (
	"errors"
	"io"
)

// Notifier publishes audit events to all subscribed observers.
type Notifier interface {
	Notify(*Event) error
}

type Publisher struct {
	observers []Observer
}

func NewPublisher(observers ...Observer) *Publisher {
	return &Publisher{
		observers: observers,
	}
}

// Notify sends the event to every observer.
// A failing observer does not prevent others from being notified; all errors are joined and returned.
func (p *Publisher) Notify(event *Event) error {
	var errs []error
	for _, observer := range p.observers {
		if err := observer.WriteEvent(event); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Close closes all observers that implement io.Closer (e.g. AsyncObserver), flushing their queues.
func (p *Publisher) Close() error {
	var errs []error
	for _, observer := range p.observers {
		if c, ok := observer.(io.Closer); ok {
			if err := c.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}
