package domain_test

import (
	"errors"
	"testing"

	"github.com/jeffersonbrasilino/ddgo/domain"
)

func TestRequiredValidator(t *testing.T) {
	t.Run("fails when string field is empty", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Name string `domainValidator:"required"`
		}
		result, err := v.Validate(&s{Name: ""})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Name"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("required should fail for empty string, got %v", result)
		}
	})

	t.Run("passes when string field is non-empty", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Name string `domainValidator:"required"`
		}
		result, err := v.Validate(&s{Name: "x"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("required should pass for non-empty string, got %v", result)
		}
	})

	t.Run("fails when slice field is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Tags []string `domainValidator:"required"`
		}
		result, err := v.Validate(&s{Tags: nil})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Tags"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("required should fail for nil slice, got %v", result)
		}
	})

	t.Run("passes when slice field is non-nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Tags []string `domainValidator:"required"`
		}
		result, err := v.Validate(&s{Tags: []string{}})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("required should pass for non-nil slice, got %v", result)
		}
	})

	t.Run("fails when map field is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Meta map[string]int `domainValidator:"required"`
		}
		result, err := v.Validate(&s{Meta: nil})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Meta"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("required should fail for nil map, got %v", result)
		}
	})

	t.Run("fails when interface field is nil", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Data interface{} `domainValidator:"required"`
		}
		result, err := v.Validate(&s{Data: nil})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Data"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("required should fail for nil interface, got %v", result)
		}
	})

	t.Run("fails when nested struct field is zero value", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type inner struct {
			X string `domainValidator:"required"`
		}
		type s struct {
			Inner inner
		}
		result, err := v.Validate(&s{Inner: inner{}})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Inner.X"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "required" {
			t.Errorf("required should fail for zero nested struct field, got %v", result)
		}
	})

	t.Run("returns error when required has a parameter", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Name string `domainValidator:"required=x"`
		}
		result, err := v.Validate(&s{Name: "x"})
		if !errors.Is(err, domain.ErrInvalidParameter) {
			t.Errorf("Validate should return ErrInvalidParameter for required=x, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})
}

func TestGteValidator(t *testing.T) {
	t.Run("fails when value length is less than param", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"gte=3"`
		}
		result, err := v.Validate(&s{Value: "ab"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Value"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "gte=3" {
			t.Errorf("gte=3 should fail for len 2, got %v", result)
		}
	})

	t.Run("passes when value length is greater or equal to param", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"gte=2"`
		}
		result, err := v.Validate(&s{Value: "ab"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("gte=2 should pass for len 2, got %v", result)
		}
	})

	t.Run("returns error when param is not a valid integer", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"gte=abc"`
		}
		result, err := v.Validate(&s{Value: "anything"})
		if !errors.Is(err, domain.ErrInvalidParameter) {
			t.Errorf("Validate should return ErrInvalidParameter for gte=abc, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})

	t.Run("returns error when field kind has no length", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Count int `domainValidator:"gte=2"`
		}
		result, err := v.Validate(&s{Count: 5})
		if !errors.Is(err, domain.ErrUnsupportedFieldKind) {
			t.Errorf("Validate should return ErrUnsupportedFieldKind for gte on int, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})
}

func TestLteValidator(t *testing.T) {
	t.Run("fails when value length is greater than param", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"lte=2"`
		}
		result, err := v.Validate(&s{Value: "abc"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Value"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "lte=2" {
			t.Errorf("lte=2 should fail for len 3, got %v", result)
		}
	})

	t.Run("passes when value length is less or equal to param", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"lte=3"`
		}
		result, err := v.Validate(&s{Value: "ab"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("lte=3 should pass for len 2, got %v", result)
		}
	})

	t.Run("returns error when param is not a valid integer", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"lte=xyz"`
		}
		result, err := v.Validate(&s{Value: "a"})
		if !errors.Is(err, domain.ErrInvalidParameter) {
			t.Errorf("Validate should return ErrInvalidParameter for lte=xyz, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})

	t.Run("returns error when field kind has no length", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Count int `domainValidator:"lte=2"`
		}
		result, err := v.Validate(&s{Count: 1})
		if !errors.Is(err, domain.ErrUnsupportedFieldKind) {
			t.Errorf("Validate should return ErrUnsupportedFieldKind for lte on int, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})
}

func TestLenValidator(t *testing.T) {
	t.Run("passes when value length equals param", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"len=3"`
		}
		result, err := v.Validate(&s{Value: "abc"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("len=3 should pass for len 3, got %v", result)
		}
	})

	t.Run("fails when value length differs from param", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"len=3"`
		}
		result, err := v.Validate(&s{Value: "ab"})
		if err != nil {
			t.Fatalf("Validate should not return error: %v", err)
		}
		r, ok := result["Value"]
		if len(result) != 1 || !ok || len(r.FailedValidators) != 1 || r.FailedValidators[0] != "len=3" {
			t.Errorf("len=3 should fail for len 2, got %v", result)
		}
	})

	t.Run("returns error when param is not a valid integer", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Value string `domainValidator:"len=invalid"`
		}
		result, err := v.Validate(&s{Value: "x"})
		if !errors.Is(err, domain.ErrInvalidParameter) {
			t.Errorf("Validate should return ErrInvalidParameter for len=invalid, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})

	t.Run("returns error when field kind has no length", func(t *testing.T) {
		t.Parallel()
		v := domain.NewValidator()
		type s struct {
			Count int `domainValidator:"len=2"`
		}
		result, err := v.Validate(&s{Count: 2})
		if !errors.Is(err, domain.ErrUnsupportedFieldKind) {
			t.Errorf("Validate should return ErrUnsupportedFieldKind for len on int, got %v", err)
		}
		if result != nil {
			t.Errorf("Validate should return nil result on error, got %v", result)
		}
	})
}
