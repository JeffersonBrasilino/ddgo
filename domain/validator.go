// Package domain provides domain validation, entities, and error types.
//
// Intent: Centralize domain building blocks (validation by struct tags,
// entities with identity, aggregate roots, and standardized errors).
// Objective: Enable consistent validation and error handling across
// the application without depending on infrastructure details.
package domain

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	mainTagName = "domainValidator"
)

var (
	// ErrNotAStruct is returned by Validate when the input is not a struct
	// or a pointer to a struct.
	ErrNotAStruct = errors.New("input must be a struct or a pointer to a struct")

	// ErrUnknownValidator is returned when a `domainValidator` tag
	// references a validator that is not registered.
	ErrUnknownValidator = errors.New("unknown validator")

	// ErrInvalidParameter is returned when a validator parameter cannot be
	// used (e.g. a non-integer parameter on gte, lte or len).
	ErrInvalidParameter = errors.New("invalid validator parameter")

	// ErrUnsupportedFieldKind is returned when a validator does not support
	// the kind of the tagged field (e.g. gte on an int field).
	ErrUnsupportedFieldKind = errors.New("unsupported field kind for validator")

	// validators maps a tag name to the factory that binds its parameter
	// into a runnable validation function.
	validators = map[string]func(fieldType reflect.Type, params string) (validateFunc, error){
		"required": requiredValidator,
		"gte":      gteValidator,
		"lte":      lteValidator,
		"len":      lenValidator,
	}
)

// validateFunc reports whether a reflected value satisfies one validation
// rule. The present flag is false when the field path crosses a nil
// pointer, meaning there is no value to validate.
type validateFunc func(value reflect.Value, present bool) bool

// ValidateResult holds the outcome of validating a single field. A field
// only appears in the result map when it fails, and FailedValidators
// preserves the tag declaration order.
type ValidateResult struct {
	FailedValidators []string
}

// boundValidator pairs a full tag (name and parameters) with its runnable
// validation function.
type boundValidator struct {
	tag string
	fn  validateFunc
}

// Validator validates structs based on `domainValidator` struct tags,
// caching the built validation schema per type.
//
// Intent: Perform validation of DTOs and domain inputs using reflection
// with a cached schema to reduce repeated reflection cost.
// Objective: Allow consumers (and dependency injection frameworks) to own
// a validator instance with its own cache state.
type Validator struct {
	mu      sync.RWMutex
	schemas map[reflect.Type][]validationStep
}

// validationStep is one cached field entry: the dotted field path, how to
// reach the field by index, and the validators bound to it.
type validationStep struct {
	fieldName        string
	indexField       []int
	boundValidators  []boundValidator
	valueExtractFunc func(reflect.Value) (reflect.Value, bool)
}

// parentFieldData tracks the index path and name path of a nested struct
// field while the validation schema is being built.
type parentFieldData struct {
	parentIndexPosition []int
	parentFieldName     []string
}

// NewValidator returns a new Validator with an empty schema cache.
//
// Intent: provide a constructor for Validator instances so consumers (and
// dependency injection frameworks) own their validator and its cache.
//
// Returns: pointer to a new Validator.
func NewValidator() *Validator {
	return &Validator{
		schemas: make(map[reflect.Type][]validationStep),
	}
}

// Validate inspects the provided struct value and runs the configured domain
// validators declared on struct fields using the `domainValidator` tag.
//
// Intent: perform validation of a DTO or domain input using reflection and a
// cached validation schema to reduce repeated reflection cost.
//
// Parameters:
//   - data: a struct or a pointer to a struct. When nil (including a typed
//     nil pointer), the function returns (nil, nil).
//
// Returns:
//   - map[string]ValidateResult: map keyed by dotted field path
//     (e.g. "Address.Street") for the failing fields. FailedValidators
//     preserves the tag declaration order. Empty map means no validation
//     failures.
//   - error: non-nil when the input is not a struct (ErrNotAStruct) or the
//     schema could not be built (e.g. ErrUnknownValidator,
//     ErrInvalidParameter, ErrUnsupportedFieldKind).
//
// Behavior:
//   - The function builds or retrieves a cached schema for the concrete type
//     and executes each configured validator. It does not short-circuit; all
//     configured validators for each field are evaluated and any failures are
//     reported in the returned map.
//   - A field whose path crosses a nil pointer has no value: required fails
//     and the length validators (gte, lte, len) are skipped.
//
// Available validators: required, gte, lte, len. The gte, lte and len
// validators are length-based, and len=N means the length must be exactly N.
//
// Example usage:
//
//	type createUserDTO struct {
//	    Name  string `domainValidator:"required"`
//	    Email string `domainValidator:"gte=5"`
//	}
//
//	v := domain.NewValidator()
//	dto := createUserDTO{Name: "", Email: "ab"}
//	result, err := v.Validate(&dto)
//	// result will contain validation failures for Name and Email.
func (d *Validator) Validate(data any) (map[string]ValidateResult, error) {
	if data == nil {
		return nil, nil
	}

	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		if dataValue.IsNil() {
			return nil, nil
		}
		dataValue = dataValue.Elem()
	}
	if dataValue.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}

	validationSchema, err := d.getValidationSchema(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]ValidateResult)
	for _, step := range validationSchema {
		value, present := step.valueExtractFunc(dataValue)
		var failedValidatorsTag []string
		for _, bound := range step.boundValidators {
			if bound.fn(value, present) {
				continue
			}
			failedValidatorsTag = append(failedValidatorsTag, bound.tag)
		}

		if len(failedValidatorsTag) == 0 {
			continue
		}

		result[step.fieldName] = ValidateResult{
			FailedValidators: failedValidatorsTag,
		}
	}

	return result, nil
}

// getValidationSchema returns the cached validation schema for the data
// type, building and caching it on first use.
//
// Parameters:
//   - data: value whose concrete type defines the cache key; pointer types
//     are unwrapped to their element type.
//
// Returns:
//   - []validationStep: cached (or newly built) validation steps.
//   - error: non-nil if the schema could not be built.
func (d *Validator) getValidationSchema(data any) ([]validationStep, error) {
	key := reflect.TypeOf(data)
	if key.Kind() == reflect.Ptr {
		key = key.Elem()
	}

	d.mu.RLock()
	schema, ok := d.schemas[key]
	d.mu.RUnlock()
	if ok {
		return schema, nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if schema, ok := d.schemas[key]; ok {
		return schema, nil
	}

	schema, err := d.makeValidationSchema(key, nil)
	if err != nil {
		return nil, err
	}
	d.schemas[key] = schema
	return schema, nil
}

// makeValidationSchema walks the struct fields recursively and builds a
// validation step for every field carrying a `domainValidator` tag. The
// walk is type-driven, so nested pointer fields are descended regardless
// of the value being validated.
//
// Parameters:
//   - t: struct type being walked.
//   - parentField: index and name path of the parent field for nested
//     structs, or nil at the top level.
//
// Returns:
//   - []validationStep: steps for the tagged fields found at this level and
//     nested levels.
//   - error: non-nil if a tag references an unknown validator or an
//     incompatible field kind.
func (d *Validator) makeValidationSchema(
	t reflect.Type,
	parentField *parentFieldData,
) ([]validationStep, error) {
	prefixPath := []int{}
	prefixName := []string{}
	if parentField != nil {
		prefixPath = parentField.parentIndexPosition
		prefixName = parentField.parentFieldName
	}

	result := []validationStep{}
	for i := 0; i < t.NumField(); i++ {
		fieldInfo := t.Field(i)
		fieldType := fieldInfo.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		currentPath := make([]int, len(prefixPath), len(prefixPath)+1)
		copy(currentPath, prefixPath)
		currentPath = append(currentPath, i)

		if fieldType.Kind() == reflect.Struct {
			childSteps, err := d.makeValidationSchema(fieldType, &parentFieldData{
				parentIndexPosition: currentPath,
				parentFieldName:     appendName(prefixName, fieldInfo.Name),
			})
			if err != nil {
				return nil, err
			}

			result = append(result, childSteps...)
			continue
		}

		validatorsTags := fieldInfo.Tag.Get(mainTagName)
		if validatorsTags == "" {
			continue
		}
		boundValidators, errValidators := d.makeValidationsFuncs(fieldType, validatorsTags)
		if errValidators != nil {
			return nil, fmt.Errorf("[domain-validator] %w", errValidators)
		}

		step := validationStep{
			fieldName:        strings.Join(appendName(prefixName, fieldInfo.Name), "."),
			indexField:       currentPath,
			boundValidators:  boundValidators,
			valueExtractFunc: d.makeStructDataAccessor(currentPath),
		}
		result = append(result, step)
	}
	return result, nil
}

// appendName appends name to names in a new slice, so the parent name path
// is never aliased by sibling appends.
func appendName(names []string, name string) []string {
	updated := make([]string, len(names)+1)
	copy(updated, names)
	updated[len(names)] = name
	return updated
}

// makeValidationsFuncs parses a comma-separated tag value and binds each
// validator name with its parameters into runnable validation functions,
// preserving the tag declaration order.
//
// Parameters:
//   - fieldType: type of the field the tags are bound to; used to reject
//     validators that do not support the field kind.
//   - validationTags: comma-separated tag value, e.g. "required,gte=2".
//
// Returns:
//   - []boundValidator: validation functions bound to their full tag (name
//     and parameters), in declaration order.
//   - error: non-nil if the tag references an unknown validator, an invalid
//     parameter, or an unsupported field kind.
func (d *Validator) makeValidationsFuncs(
	fieldType reflect.Type,
	validationTags string,
) ([]boundValidator, error) {
	validatorsNames := strings.Split(validationTags, ",")
	bound := make([]boundValidator, 0, len(validatorsNames))
	for _, tag := range validatorsNames {
		tagName, tagParams, _ := strings.Cut(tag, "=")
		factory, ok := validators[tagName]
		if !ok {
			return nil, fmt.Errorf(
				"validator %q does not exist: %w", tagName, ErrUnknownValidator,
			)
		}

		validatorFunc, err := factory(fieldType, tagParams)
		if err != nil {
			return nil, err
		}
		bound = append(bound, boundValidator{tag: tag, fn: validatorFunc})
	}
	return bound, nil
}

// makeStructDataAccessor builds the function that extracts a field value by
// its index path from a struct value at validation time.
//
// Parameters:
//   - fieldAddress: index path from the root struct to the field.
//
// Returns: a function that receives the root struct value and returns the
// field value at the given path plus a present flag, false when the path
// crosses a nil pointer (no value to validate).
func (d *Validator) makeStructDataAccessor(
	fieldAddress []int,
) func(dataStruct reflect.Value) (reflect.Value, bool) {
	return func(dataStruct reflect.Value) (reflect.Value, bool) {
		v := dataStruct
		for _, idx := range fieldAddress {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					return reflect.Value{}, false
				}
				v = v.Elem()
			}
			v = v.Field(idx)
		}
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}, false
			}
			v = v.Elem()
		}
		return v, true
	}
}
