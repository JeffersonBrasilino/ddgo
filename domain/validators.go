// Validators used by the domain validator (tag-based validation).
//
// Intent: Provide required, gte, lte, and len checks for struct fields
// when using the domainValidator tag.
// Objective: Keep validation logic in one place and reusable by the
// validator without exporting these helpers.

package domain

import (
	"fmt"
	"reflect"
	"strconv"
)

// isNilableKind reports whether the value kind can hold a nil reference
// (interface, chan, func, map, pointer, or slice).
//
// Parameters:
//   - value: reflected value to inspect.
//
// Returns: true when the kind can hold a nil reference.
func isNilableKind(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Interface, reflect.Chan, reflect.Func,
		reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	default:
		return false
	}
}

// requireLenKind reports an error when the validator does not support the
// field kind: gte, lte and len operate on values with length.
//
// Parameters:
//   - validatorName: validator name for the error message.
//   - fieldType: type of the field being validated.
//
// Returns: nil when the kind has length; a wrapped ErrUnsupportedFieldKind
// otherwise.
func requireLenKind(validatorName string, fieldType reflect.Type) error {
	switch fieldType.Kind() {
	case reflect.String, reflect.Slice, reflect.Map,
		reflect.Array, reflect.Chan:
		return nil
	default:
		return fmt.Errorf(
			"validator %s requires a field with length (string, slice, map, array or chan), got %s: %w",
			validatorName, fieldType.Kind(), ErrUnsupportedFieldKind,
		)
	}
}

// requiredValidator fails when the field is absent (nil pointer in its
// path), a nil complex value, or the zero value of its kind.
//
// Parameters:
//   - fieldType: type of the field being validated; unused by this validator.
//   - param: this validator takes no parameters; any value is an error.
//
// Returns: a validateFunc that reports whether the value is present.
func requiredValidator(fieldType reflect.Type, param string) (validateFunc, error) {
	if param != "" {
		return nil, fmt.Errorf(
			"validator required does not accept parameters, got %q: %w",
			param, ErrInvalidParameter,
		)
	}
	return func(value reflect.Value, present bool) bool {
		if !present {
			return false
		}
		if isNilableKind(value) && value.IsNil() {
			return false
		}
		if value.IsZero() {
			return false
		}
		return true
	}, nil
}

// gteValidator fails when the value length is lower than the parameter.
//
// Parameters:
//   - fieldType: type of the field being validated; must have length.
//   - param: integer string with the minimum accepted length.
//
// Returns: a validateFunc that reports whether the value length is greater
// than or equal to the parameter.
func gteValidator(fieldType reflect.Type, param string) (validateFunc, error) {
	if err := requireLenKind("gte", fieldType); err != nil {
		return nil, err
	}
	want, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"validator gte expects an integer parameter, got %q: %w",
			param, ErrInvalidParameter,
		)
	}
	return func(value reflect.Value, present bool) bool {
		if !present {
			return true
		}
		return value.Len() >= want
	}, nil
}

// lteValidator fails when the value length is greater than the parameter.
//
// Parameters:
//   - fieldType: type of the field being validated; must have length.
//   - param: integer string with the maximum accepted length.
//
// Returns: a validateFunc that reports whether the value length is lower
// than or equal to the parameter.
func lteValidator(fieldType reflect.Type, param string) (validateFunc, error) {
	if err := requireLenKind("lte", fieldType); err != nil {
		return nil, err
	}
	want, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"validator lte expects an integer parameter, got %q: %w",
			param, ErrInvalidParameter,
		)
	}
	return func(value reflect.Value, present bool) bool {
		if !present {
			return true
		}
		return value.Len() <= want
	}, nil
}

// lenValidator fails when the value length differs from the parameter:
// len=N means the length must be exactly N.
//
// Parameters:
//   - fieldType: type of the field being validated; must have length.
//   - param: integer string with the exact accepted length.
//
// Returns: a validateFunc that reports whether the value length equals the
// parameter.
func lenValidator(fieldType reflect.Type, param string) (validateFunc, error) {
	if err := requireLenKind("len", fieldType); err != nil {
		return nil, err
	}
	want, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"validator len expects an integer parameter, got %q: %w",
			param, ErrInvalidParameter,
		)
	}
	return func(value reflect.Value, present bool) bool {
		if !present {
			return true
		}
		return value.Len() == want
	}, nil
}
