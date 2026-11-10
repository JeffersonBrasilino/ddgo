package domain_test

import (
	"testing"
	"time"

	"github.com/jeffersonbrasilino/ddgo/domain"
)

type fakeEvent struct{}

func (f *fakeEvent) Headers() map[string]string {
	return map[string]string{"type": "fake"}
}

func (f *fakeEvent) OccurredOn() time.Time {
	return time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
}

func (f *fakeEvent) Payload() any {
	return "ok"
}

type order struct {
	domain.Aggregate
}

func TestNewAggregateRoot(t *testing.T) {
	t.Run("returns aggregate with given id", func(t *testing.T) {
		t.Parallel()
		id := "550e8400-e29b-41d4-a716-446655440000"
		ar := domain.NewAggregateRoot(id)
		if ar.ID() != id {
			t.Errorf("aggregate id should be %q, got %q", id, ar.ID())
		}
	})

	t.Run("returns aggregate with empty id when empty string is passed", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("")
		if ar.ID() != "" {
			t.Errorf("aggregate id should be empty, got %q", ar.ID())
		}
	})

	t.Run("returns aggregate with no pending events", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		if got := ar.Events(); len(got) != 0 {
			t.Errorf("new aggregate should have no pending events, got %d", len(got))
		}
	})
}

func TestAggregate_AddEvent(t *testing.T) {
	t.Run("records the event on the pending list", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		ev := &fakeEvent{}
		ar.AddEvent(ev)
		got := ar.Events()
		if len(got) != 1 {
			t.Fatalf("Events should return exactly 1 event, got %d", len(got))
		}
		if got[0] != ev {
			t.Error("Events should return the same event added")
		}
	})

	t.Run("silently ignores nil event", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		ar.AddEvent(nil)
		if got := ar.Events(); len(got) != 0 {
			t.Errorf("nil event should not be recorded, got %d events", len(got))
		}
	})

	t.Run("records typed-nil event as-is (documented contract: callers must not pass it)", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		// Design decision: a typed-nil (*fakeEvent)(nil) is not (nil, nil)
		// as an interface, so the nil guard does not catch it. Rejecting it
		// would require reflection, judged not worth the cost — the contract
		// is documented on AddEvent instead.
		var typedNil *fakeEvent
		ar.AddEvent(typedNil)
		got := ar.Events()
		if len(got) != 1 {
			t.Fatalf("typed-nil event should be recorded as-is, got %d events", len(got))
		}
		if got[0] != typedNil {
			t.Error("Events should return the same typed-nil event added")
		}
	})

	t.Run("preserves insertion order of multiple events", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		first, second, third := &fakeEvent{}, &fakeEvent{}, &fakeEvent{}
		ar.AddEvent(first)
		ar.AddEvent(second)
		ar.AddEvent(third)
		got := ar.Events()
		if len(got) != 3 {
			t.Fatalf("Events should return exactly 3 events, got %d", len(got))
		}
		if got[0] != first || got[1] != second || got[2] != third {
			t.Error("Events should preserve insertion order")
		}
	})
}

func TestAggregate_Events(t *testing.T) {
	t.Run("returns a defensive copy so mutating the result does not affect the aggregate", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		ar.AddEvent(&fakeEvent{})
		got := ar.Events()
		got[0] = nil
		if again := ar.Events(); len(again) != 1 || again[0] == nil {
			t.Error("mutating the Events result should not affect the aggregate pending list")
		}
	})

	t.Run("exposes event data through the Event interface", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		ar.AddEvent(&fakeEvent{})
		got := ar.Events()
		if len(got) != 1 {
			t.Fatalf("Events should return exactly 1 event, got %d", len(got))
		}
		if _, ok := got[0].Headers()["type"]; !ok {
			t.Error("event headers should be accessible through the interface")
		}
		if got[0].OccurredOn().IsZero() {
			t.Error("event occurred-on should be a valid time, got zero")
		}
		if got[0].Payload() != "ok" {
			t.Errorf("event payload should be %q, got %v", "ok", got[0].Payload())
		}
	})
}

func TestAggregate_ClearEvents(t *testing.T) {
	t.Run("discards all pending events", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		ar.AddEvent(&fakeEvent{})
		ar.AddEvent(&fakeEvent{})
		ar.ClearEvents()
		if got := ar.Events(); len(got) != 0 {
			t.Errorf("ClearEvents should discard all events, got %d", len(got))
		}
	})

	t.Run("is safe on an aggregate without events", func(t *testing.T) {
		t.Parallel()
		ar := domain.NewAggregateRoot("mock")
		ar.ClearEvents()
		if got := ar.Events(); len(got) != 0 {
			t.Errorf("ClearEvents on empty aggregate should keep it empty, got %d", len(got))
		}
	})
}

func TestAggregate_Embedding(t *testing.T) {
	t.Run("records events on a concrete root that embeds Aggregate", func(t *testing.T) {
		t.Parallel()
		o := &order{Aggregate: domain.NewAggregateRoot("order-1")}
		o.AddEvent(&fakeEvent{})
		if o.ID() != "order-1" {
			t.Errorf("embedded ID() should return %q, got %q", "order-1", o.ID())
		}
		if got := o.Events(); len(got) != 1 {
			t.Errorf("embedded AddEvent should record 1 event, got %d", len(got))
		}
	})
}
