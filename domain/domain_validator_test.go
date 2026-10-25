package domain_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/jeffersonbrasilino/ddgo/domain"
)

func TestNewValidator(t *testing.T) {
	t.Run("returns a usable validator", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		if v == nil {
			t.Fatal("NewValidator should not return nil")
		}
		type withRequired struct {
			Name string `domainValidator:"required"`
		}
		result, err := v.Validate(&withRequired{Name: "a"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no failures for valid dto, got %v", result)
		}
	})
}

func TestValidator_Validate(t *testing.T) {
	t.Run("returns nil map and nil error when data is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		result, err := v.Validate(nil)
		if result != nil {
			t.Errorf("Validate(nil) should return nil map, got %v", result)
		}
		if err != nil {
			t.Errorf("Validate(nil) should return nil error, got %v", err)
		}
	})

	t.Run("returns nil map and nil error when data is a typed-nil pointer", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withRequired struct {
			Name string `domainValidator:"required"`
		}
		var typedNil *withRequired
		result, err := v.Validate(typedNil)
		if result != nil {
			t.Errorf("Validate(typed-nil) should return nil map, got %v", result)
		}
		if err != nil {
			t.Errorf("Validate(typed-nil) should return nil error, got %v", err)
		}
	})

	t.Run("returns error when data is not a struct", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		for _, input := range []any{"string", 42, true, 3.14} {
			result, err := v.Validate(input)
			if !errors.Is(err, domain.ErrNotAStruct) {
				t.Errorf("Validate(%v) should return ErrNotAStruct, got %v", input, err)
			}
			if result != nil {
				t.Errorf("Validate(%v) should return nil result on error, got %v", input, result)
			}
		}
	})

	t.Run("returns empty map when struct has no domainValidator tags", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type noTags struct {
			Name string
		}
		dto := &noTags{Name: "x"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result for struct without tags, got %v", result)
		}
	})

	t.Run("returns failure when required field is empty", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withRequired struct {
			Name string `domainValidator:"required"`
		}
		dto := &withRequired{Name: ""}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected one failure, got %d: %v", len(result), result)
		}
		if r, ok := result["Name"]; !ok {
			t.Errorf("expected failure for field Name, got keys: %v", result)
		} else if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("expected FailedValidators [required], got %v", r.FailedValidators)
		}
	})

	t.Run("returns empty map when required field is set", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withRequired struct {
			Name string `domainValidator:"required"`
		}
		dto := &withRequired{Name: "ok"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no failures when required is set, got %v", result)
		}
	})

	t.Run("returns failure when gte validator fails", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withGte struct {
			Value string `domainValidator:"gte=3"`
		}
		dto := &withGte{Value: "ab"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected one failure for gte=3 with len 2, got %v", result)
		}
		if r, ok := result["Value"]; !ok {
			t.Errorf("expected failure for Value, got %v", result)
		} else if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "gte=3" {
			t.Errorf("expected FailedValidators [gte=3], got %v", r.FailedValidators)
		}
	})

	t.Run("returns empty map when gte validator passes", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withGte struct {
			Value string `domainValidator:"gte=2"`
		}
		dto := &withGte{Value: "ab"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no failures when gte passes, got %v", result)
		}
	})

	t.Run("returns failure when lte validator fails", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withLte struct {
			Value string `domainValidator:"lte=2"`
		}
		dto := &withLte{Value: "abc"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected one failure for lte=2 with len 3, got %v", result)
		}
		if r, ok := result["Value"]; !ok {
			t.Errorf("expected failure for Value, got %v", result)
		} else if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "lte=2" {
			t.Errorf("expected FailedValidators [lte=2], got %v", r.FailedValidators)
		}
	})

	t.Run("returns failure when len validator fails", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withLen struct {
			Value string `domainValidator:"len=3"`
		}
		dto := &withLen{Value: "ab"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected one failure for len=3 with value len 2, got %v", result)
		}
		if r, ok := result["Value"]; !ok {
			t.Errorf("expected failure for Value, got %v", result)
		} else if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "len=3" {
			t.Errorf("expected FailedValidators [len=3], got %v", r.FailedValidators)
		}
	})

	t.Run("returns empty map when len validator passes", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withLen struct {
			Value string `domainValidator:"len=3"`
		}
		dto := &withLen{Value: "abc"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no failures when len passes, got %v", result)
		}
	})

	t.Run("returns error when validator name does not exist", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type invalidValidator struct {
			Name string `domainValidator:"nonexistent"`
		}
		dto := &invalidValidator{Name: "x"}
		result, err := v.Validate(dto)
		if !errors.Is(err, domain.ErrUnknownValidator) {
			t.Errorf("Validate should return ErrUnknownValidator, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})

	t.Run("returns error when gte is used on a field without length", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withIntGte struct {
			Age int `domainValidator:"gte=18"`
		}
		dto := &withIntGte{Age: 17}
		result, err := v.Validate(dto)
		if !errors.Is(err, domain.ErrUnsupportedFieldKind) {
			t.Errorf("Validate should return ErrUnsupportedFieldKind, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})

	t.Run("preserves tag declaration order in FailedValidators", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type multi struct {
			Value string `domainValidator:"required,gte=3"`
		}
		dto := &multi{Value: ""}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Value"]
		if !ok {
			t.Fatalf("expected failure for Value, got %v", result)
		}
		if len(r.FailedValidators) != 2 ||
			r.FailedValidators[0] != "required" || r.FailedValidators[1] != "gte=3" {
			t.Errorf("expected FailedValidators [required gte=3] in declaration order, got %v", r.FailedValidators)
		}
	})

	t.Run("validates nested struct fields", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type nested struct {
			Street string `domainValidator:"required"`
		}
		type withNested struct {
			Name    string `domainValidator:"required"`
			Address nested
		}
		dto := &withNested{Name: "ok", Address: nested{Street: ""}}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected one failure for nested Street, got %v", result)
		}
		if _, ok := result["Address.Street"]; !ok {
			t.Errorf("expected failure at Address.Street, got keys: %v", result)
		}
	})

	t.Run("accepts pointer to struct", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withRequired struct {
			Name string `domainValidator:"required"`
		}
		dto := &withRequired{Name: "a"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no failures for valid pointer, got %v", result)
		}
	})

	t.Run("uses cached schema on second validate of same type", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type withRequired struct {
			Name string `domainValidator:"required"`
		}
		dto1 := &withRequired{Name: "first"}
		dto2 := &withRequired{Name: "second"}
		if _, err := v.Validate(dto1); err != nil {
			t.Fatalf("first Validate should not return error: %v", err)
		}
		result2, err := v.Validate(dto2)
		if err != nil {
			t.Fatalf("second Validate should not return error: %v", err)
		}
		if len(result2) != 0 {
			t.Errorf("second validate should succeed: %v", result2)
		}
	})

	t.Run("multiple validators on same field all run", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type multi struct {
			Value string `domainValidator:"required,gte=2"`
		}
		dtoEmpty := &multi{Value: ""}
		resultEmpty, _ := v.Validate(dtoEmpty)
		r := resultEmpty["Value"]
		if len(r.FailedValidators) != 2 ||
			r.FailedValidators[0] != "required" || r.FailedValidators[1] != "gte=2" {
			t.Errorf("expected FailedValidators [required gte=2], got %v", r.FailedValidators)
		}

		dtoShort := &multi{Value: "a"}
		resultShort, _ := v.Validate(dtoShort)
		r = resultShort["Value"]
		if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "gte=2" {
			t.Errorf("expected FailedValidators [gte=2], got %v", r.FailedValidators)
		}

		dtoOk := &multi{Value: "ab"}
		resultOk, _ := v.Validate(dtoOk)
		if len(resultOk) != 0 {
			t.Errorf("expected no failures for valid Value, got %v", resultOk)
		}
	})

	t.Run("field with empty domainValidator tag is skipped", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type emptyTag struct {
			Name string `domainValidator:""`
		}
		dto := &emptyTag{Name: "x"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no validation for empty tag, got %v", result)
		}
	})

	t.Run("validates nested struct via pointer field", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type inner struct {
			Street string `domainValidator:"required"`
		}
		type withPtrNested struct {
			Address *inner
		}
		dto := &withPtrNested{Address: &inner{Street: ""}}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected one failure for nested pointer struct, got %v", result)
		}
		if _, ok := result["Address.Street"]; !ok {
			t.Errorf("expected failure at Address.Street, got keys: %v", result)
		}
	})

	t.Run("validates nested struct via pointer field when valid", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type inner struct {
			Street string `domainValidator:"required"`
		}
		type withPtrNested struct {
			Address *inner
		}
		dto := &withPtrNested{Address: &inner{Street: "valid"}}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no failures when nested pointer struct is valid, got %v", result)
		}
	})

	t.Run("returns error when nested field has unknown validator", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type inner struct {
			Street string `domainValidator:"nonexistent"`
		}
		type outer struct {
			Name    string `domainValidator:"required"`
			Address inner
		}
		dto := &outer{Name: "x"}
		result, err := v.Validate(dto)
		if !errors.Is(err, domain.ErrUnknownValidator) {
			t.Errorf("Validate should return ErrUnknownValidator for nested field, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})

	t.Run("does not panic and reports required failure when nested pointer is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type inner struct {
			Street string `domainValidator:"required"`
		}
		type middle struct {
			City *inner
		}
		type root struct {
			Name   string `domainValidator:"required"`
			Family *middle
		}
		dto := &root{Name: "ok"}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Family.City.Street"]
		if !ok {
			t.Fatalf("expected failure for Family.City.Street when pointer is nil, got %v", result)
		}
		if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("expected only required to fail for nil nested pointer, got %v", r)
		}
	})

	t.Run("does not panic when a pointer in the middle of the path is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type inner struct {
			Street string `domainValidator:"required"`
		}
		type middle struct {
			City *inner
		}
		type root struct {
			Name   string `domainValidator:"required"`
			Family *middle
		}
		dto := &root{Name: "ok", Family: &middle{}}
		result, err := v.Validate(dto)
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if _, ok := result["Family.City.Street"]; !ok {
			t.Errorf("expected failure for Family.City.Street when middle pointer is nil, got %v", result)
		}
	})

	t.Run("validates nested pointer regardless of which instance is validated first", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type inner struct {
			Street string `domainValidator:"required"`
		}
		type middle struct {
			City *inner
		}
		type root struct {
			Name   string `domainValidator:"required"`
			Family *middle
		}
		nilFirst := &root{Name: "ok"}
		if _, err := v.Validate(nilFirst); err != nil {
			t.Fatalf("first Validate should not return error: %v", err)
		}
		invalidAfter := &root{Name: "ok", Family: &middle{City: &inner{Street: ""}}}
		result, err := v.Validate(invalidAfter)
		if err != nil {
			t.Fatalf("second Validate should not return error: %v", err)
		}
		if _, ok := result["Family.City.Street"]; !ok {
			t.Errorf("expected failure for Family.City.Street after cache was built from a nil-pointer instance, got %v", result)
		}
	})
}

func TestValidator_PointerToRawField(t *testing.T) {
	type profile struct {
		Bio *string `domainValidator:"required,gte=3"`
	}

	t.Run("fails required and skips gte when pointer is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		result, err := v.Validate(&profile{})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Bio"]
		if !ok {
			t.Fatalf("expected failure for Bio when pointer is nil, got %v", result)
		}
		if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("expected only required to fail for nil pointer, got %v", r.FailedValidators)
		}
	})

	t.Run("fails gte on the pointed value when pointer is set", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		bio := "ab"
		result, err := v.Validate(&profile{Bio: &bio})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Bio"]
		if !ok {
			t.Fatalf("expected failure for Bio with pointed value shorter than gte, got %v", result)
		}
		if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "gte=3" {
			t.Errorf("expected only gte=3 to fail for short pointed value, got %v", r.FailedValidators)
		}
	})

	t.Run("passes when pointed value satisfies all validators", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		bio := "abc"
		result, err := v.Validate(&profile{Bio: &bio})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected no failures for valid pointed value, got %v", result)
		}
	})

	t.Run("skips lte and len when pointer is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type doc struct {
			Title *string `domainValidator:"required,lte=5"`
			Code  *string `domainValidator:"len=3"`
		}
		result, err := v.Validate(&doc{})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Title"]
		if !ok {
			t.Fatalf("expected failure for Title when pointer is nil, got %v", result)
		}
		if len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("expected only required to fail for nil pointer, got %v", r.FailedValidators)
		}
		if _, ok := result["Code"]; ok {
			t.Errorf("len should be skipped when pointer is nil, got %v", result)
		}
	})
}

func TestValidator_ConcurrentSchemaBuild(t *testing.T) {
	t.Run("builds schema once and validates correctly under concurrency", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		var wg sync.WaitGroup
		for round := 0; round < len(concurrentDTOFactories); round++ {
			for g := 0; g < 20; g++ {
				wg.Add(1)
				go func(mkDTO func() any) {
					defer wg.Done()
					if _, err := v.Validate(mkDTO()); err != nil {
						t.Errorf("concurrent Validate should not return error: %v", err)
					}
				}(concurrentDTOFactories[round])
			}
		}
		wg.Wait()
		invalid := &concurrentDTO0{Name: ""}
		result, err := v.Validate(invalid)
		if err != nil {
			t.Fatalf("Validate after concurrency should not return error: %v", err)
		}
		if _, ok := result["Name"]; !ok {
			t.Errorf("expected failure for Name after concurrent builds, got %v", result)
		}
	})
}

// Cold types validated concurrently: each one re-opens the cache race
// window in getValidationSchema, so the double-check re-read is exercised.
type concurrentDTO0 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO1 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO2 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO3 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO4 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO5 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO6 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO7 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO8 struct {
	Name string `domainValidator:"required"`
}

type concurrentDTO9 struct {
	Name string `domainValidator:"required"`
}

var concurrentDTOFactories = []func() any{
	func() any { return &concurrentDTO0{Name: "x"} },
	func() any { return &concurrentDTO1{Name: "x"} },
	func() any { return &concurrentDTO2{Name: "x"} },
	func() any { return &concurrentDTO3{Name: "x"} },
	func() any { return &concurrentDTO4{Name: "x"} },
	func() any { return &concurrentDTO5{Name: "x"} },
	func() any { return &concurrentDTO6{Name: "x"} },
	func() any { return &concurrentDTO7{Name: "x"} },
	func() any { return &concurrentDTO8{Name: "x"} },
	func() any { return &concurrentDTO9{Name: "x"} },
}
