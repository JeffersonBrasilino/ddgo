// Legacy domain entity, superseded by package
// github.com/jeffersonbrasilino/ddgo/domain.
//
// Intent: Provide a shared identity abstraction for domain types.
// Objective: Let concrete entities embed Entity and reuse uuid handling
// without infrastructure concerns.
//
// Every exported identifier in this file is deprecated and kept only for backward
// compatibility. New code must use the domain package
// (github.com/jeffersonbrasilino/ddgo/domain).

package ddgo

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
//	    *Entity
//	}
//	var itemInstance = &item{Entity: domain.NewEntity("")}
//
// Deprecated: Use domain.Entity from github.com/jeffersonbrasilino/ddgo/domain
// instead.
type Entity struct {
	uuid string
}

// NewEntity builds a new domain entity with the given UUID.
//
// Parameters:
//   - uuid: unique identifier for the entity (may be empty).
//
// Returns: a non-nil *Entity holding the given uuid.
//
// Deprecated: There is no direct replacement yet; use the Entity type from
// github.com/jeffersonbrasilino/ddgo/domain instead.
func NewEntity(uuid string) *Entity {
	return &Entity{
		uuid: uuid,
	}
}

// Uuid returns the entity's unique identifier.
//
// Returns: the uuid string passed at construction (may be empty).
//
// Deprecated: Use the Uuid method of domain.Entity from
// github.com/jeffersonbrasilino/ddgo/domain instead.
func (entity *Entity) Uuid() string {
	return entity.uuid
}
