package domain_test

import (
	"testing"

	"github.com/jeffersonbrasilino/ddgo/domain"
)

type item struct {
	domain.Entity
}

func TestNewEntity(t *testing.T) {
	t.Run("returns entity with given id", func(t *testing.T) {
		t.Parallel()
		id := "550e8400-e29b-41d4-a716-446655440000"
		e := domain.NewEntity(id)
		if e.ID() != id {
			t.Errorf("entity id should be %q, got %q", id, e.ID())
		}
	})

	t.Run("returns entity with empty id when empty string is passed", func(t *testing.T) {
		t.Parallel()
		e := domain.NewEntity("")
		if e.ID() != "" {
			t.Errorf("entity id should be empty, got %q", e.ID())
		}
	})
}

func TestEntity_Id(t *testing.T) {
	t.Run("returns id set at creation", func(t *testing.T) {
		t.Parallel()
		id := "abc-123"
		e := domain.NewEntity(id)
		got := e.ID()
		if got != id {
			t.Errorf("ID() should return %q, got %q", id, got)
		}
	})

	t.Run("returns empty string when entity was created with empty id", func(t *testing.T) {
		t.Parallel()
		e := domain.NewEntity("")
		got := e.ID()
		if got != "" {
			t.Errorf("ID() should return empty string, got %q", got)
		}
	})
}

func TestEntity_Embedding(t *testing.T) {
	t.Run("promotes Id to concrete types that embed Entity", func(t *testing.T) {
		t.Parallel()
		it := item{Entity: domain.NewEntity("item-1")}
		if it.ID() != "item-1" {
			t.Errorf("embedded ID() should return %q, got %q", "item-1", it.ID())
		}
	})

	t.Run("returns empty id through embedding when created with empty id", func(t *testing.T) {
		t.Parallel()
		it := item{Entity: domain.NewEntity("")}
		if it.ID() != "" {
			t.Errorf("embedded ID() should return empty string, got %q", it.ID())
		}
	})
}
