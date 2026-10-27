// Domain entity base type with identity handling.
//
// Intent: Provide a shared identity abstraction for domain types.
// Objective: Let concrete entities embed Entity and reuse id handling
// without infrastructure concerns.

package domain

// Entity is the base type for domain entities with a unique identifier.
//
// Intent: Provide a common abstraction where domain rules can live,
// independent of persistence or external integrations.
// Objective: Allow embedding Entity in concrete domain types so they
// share identity handling and stay free of infrastructure concerns.
//
// Example:
//
//	type item struct {
//	    domain.Entity
//	}
//	var itemInstance = item{Entity: domain.NewEntity("item-1")}
type Entity struct {
	id string
}

// NewEntity builds a new domain entity with the given id.
//
// Parameters:
//   - id: unique identifier for the entity. The library neither generates
//     nor validates it, so callers are responsible for uniqueness; it may
//     be empty for entities whose identity is assigned later.
//
// Returns: an Entity value holding the given id.
func NewEntity(id string) Entity {
	return Entity{
		id: id,
	}
}

// ID returns the entity's unique identifier.
//
// Returns: the id string passed at construction (may be empty).
func (entity *Entity) ID() string {
	return entity.id
}
