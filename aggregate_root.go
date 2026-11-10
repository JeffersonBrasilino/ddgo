// Legacy aggregate root and domain events, superseded by package
// github.com/jeffersonbrasilino/ddgo/domain.
//
// Intent: Provide a root type that holds identity and a list of domain
// events to dispatch after persistence.
// Objective: Support event sourcing and auditing. See:
// https://martinfowler.com/bliki/DDD_Aggregate.html
//
// Every exported identifier in this file is deprecated and kept only for backward
// compatibility. New code must use the domain package
// (github.com/jeffersonbrasilino/ddgo/domain).

package ddgo

import (
	"fmt"
	"time"
)

// DomainEvent is a single occurrence in the domain, with payload and metadata.
//
// Intent: Represent something that happened for dispatch and auditing.
// Implementations must provide Payload, OcurredOn, and Uuid.
//
// Deprecated: Use domain.Event from github.com/jeffersonbrasilino/ddgo/domain
// instead.
type DomainEvent interface {
	Payload() any
	OcurredOn() time.Time
	Uuid() string
}

// AggregateRoot embeds Entity and holds domain events for later dispatch.
//
// Deprecated: Use domain.Aggregate from github.com/jeffersonbrasilino/ddgo/domain
// instead.
type AggregateRoot struct {
	*Entity
	domainEvents map[string]DomainEvent
}

// NewAggregateRoot creates an aggregate root with the given UUID.
//
// Parameters: uuid — unique identifier for the aggregate.
// Returns: a new *AggregateRoot with an empty event list.
//
// Deprecated: Use domain.NewAggregateRoot from
// github.com/jeffersonbrasilino/ddgo/domain instead.
func NewAggregateRoot(uuid string) *AggregateRoot {
	return &AggregateRoot{
		NewEntity(uuid),
		map[string]DomainEvent{},
	}
}

// DomainEvents returns the map of domain events to be dispatched.
//
// Returns: map from event UUID to DomainEvent (read-only view; do not modify).
//
// Deprecated: Use the Events method of domain.Aggregate from
// github.com/jeffersonbrasilino/ddgo/domain instead.
func (a *AggregateRoot) DomainEvents() map[string]DomainEvent {
	return a.domainEvents
}

// AddDomainEvent adds an event to the dispatch list.
//
// Parameters: event — the domain event (must not be nil).
// Returns: nil on success; error if event is nil.
// Behavior: Events are keyed by event.Uuid(); duplicate UUIDs overwrite.
//
// Deprecated: Use the AddEvent method of domain.Aggregate from
// github.com/jeffersonbrasilino/ddgo/domain instead.
func (a *AggregateRoot) AddDomainEvent(event DomainEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	a.domainEvents[event.Uuid()] = event
	return nil
}

// RemoveDomainEvent removes the event with the given UUID from the list.
//
// Parameters: uuid — the UUID of the event to remove.
// Behavior: No-op if the UUID is not present.
//
// Deprecated: superseded by the domain package
// (github.com/jeffersonbrasilino/ddgo/domain).
func (a *AggregateRoot) RemoveDomainEvent(uuid string) {
	delete(a.domainEvents, uuid)
}

// ClearEvents removes all domain events from the dispatch list.
//
// Deprecated: Use the ClearEvents method of domain.Aggregate from
// github.com/jeffersonbrasilino/ddgo/domain instead.
func (a *AggregateRoot) ClearEvents() {
	a.domainEvents = make(map[string]DomainEvent)
}
