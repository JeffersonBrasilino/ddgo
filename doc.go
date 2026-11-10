// Package ddgo provides DDD building blocks for Go applications:
// entities, aggregate roots with domain events, tag-based domain
// validation, and standardized errors.
//
// The exported identifiers declared at the root are legacy and kept only
// for backward compatibility. New code must use the focused subpackages:
//
//   - domain: entities, aggregates, events, and tag-based validation
//   - err: standardized errors matched with errors.Is
package ddgo
