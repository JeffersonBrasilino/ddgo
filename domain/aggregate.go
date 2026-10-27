// Aggregate event collection for domain aggregate roots.
//
// Intent: Give aggregate roots a single place to record domain events as
// state changes happen, so they can be published later.
// Objective: Keep event recording decoupled from persistence and from the
// concrete aggregate implementations that embed it.

package domain

import (
	"slices"
	"time"
)

// Event represents a domain event recorded by an aggregate root.
//
// Intent: Model something that happened in the domain and carry the data
// needed to publish or persist it.
type Event interface {
	// Headers returns the event metadata as a string map.
	Headers() map[string]string

	// OccurredOn returns the moment the event occurred.
	OccurredOn() time.Time

	// Payload returns the event data.
	Payload() any
}

// Aggregate is the base type for domain aggregate roots with event recording.
//
// Intent: Extend Entity with a pending domain event list so state changes
// can be collected and dispatched as a unit.
// Objective: Allow concrete roots to embed Aggregate and record events
// without knowing the publishing mechanism.
type Aggregate struct {
	Entity
	events []Event
}

// NewAggregateRoot builds a new aggregate root with an empty event list.
//
// Parameters:
//   - id: unique identifier for the aggregate root.
//
// Returns: an Aggregate whose underlying Entity holds the given id and
// with no recorded events.
func NewAggregateRoot(id string) Aggregate {
	return Aggregate{
		Entity{id: id},
		make([]Event, 0),
	}
}

// AddEvent records a domain event on the aggregate's pending list.
//
// Parameters:
//   - event: domain event to append; a nil event is silently ignored.
//     Callers must not pass a typed-nil value (e.g. a nil *orderCreated):
//     it is not detected as nil and gets recorded as-is, which may panic
//     the consumer that reads it.
//
// Behavior: mutates the aggregate's event list in place; it does not
// publish the event.
func (a *Aggregate) AddEvent(event Event) {

	if event == nil {
		return
	}

	a.events = append(a.events, event)
}

// Events returns a defensive copy of the aggregate's pending domain events.
//
// Returns: the recorded events in insertion order; mutating the result
// does not affect the aggregate.
func (a *Aggregate) Events() []Event {
	return slices.Clone(a.events)
}

// ClearEvents discards all pending domain events from the aggregate.
//
// Behavior: replaces the pending event list with an empty slice.
func (a *Aggregate) ClearEvents() {
	a.events = []Event{}
}
