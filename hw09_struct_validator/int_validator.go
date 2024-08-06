package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// IntValidator is a Validator with validation rules for int type.
//
// Supported validation rules:
//
//	min:number      - value must not be less than the number
//	max:number      - value must not be great than the number
//	in:n1,n2,...,nN - value must be in the set of numbers {n1,n2,...,nN}
//
// Supports union (|) of rules.
// Example: `min:24|max:42` - value must not be at be less than 24
// and not great than 42 at the same time.
func IntValidator() Validator {
	return intValidator{
		validatorRuleMatcher: validatorRuleMatcher{
			unionSep: "|",
			ruleSep:  ":",
		},
	}
}

type intValidator struct {
	validatorRuleMatcher
}

var (
	ErrIntMin = errors.New("value less than min")
	ErrIntMax = errors.New("value great than max")
	ErrIntIn  = errors.New("value not in the set")
)

// String returns validator name.
func (r intValidator) String() string {
	return "intValidator"
}

// Supports returns true if fieldType is type of int.
func (r intValidator) Supports(fieldType reflect.Type) bool {
	return fieldType.Kind() == reflect.Int
}

func (r intValidator) Kind() reflect.Kind {
	return reflect.Int
}

// ValidatorsFor returns slice of value validators for provided rules.
// Returns ErrTypeNotSupported if fieldType is not supported by validator.
func (r intValidator) ValidatorsFor(fieldType reflect.Type, rules string) ([]ValueValidatorFn, error) {
	// Check if validator supports specified struct field.
	if !r.Supports(fieldType) {
		return nil, ErrTypeNotSupported
	}

	ruleMap := map[string]genValidatorFn{
		"min": r.validatorMin,
		"max": r.validatorMax,
		"in":  r.validatorIn,
	}

	return r.validatorsFor(rules, ruleMap)
}

// validatorMin is a generator of "min"-rule validator for ruleCond.
// ValueValidatorFn accepts fieldValue of Kind int.
func (r intValidator) validatorMin(ruleCond string) (ValueValidatorFn, error) {
	minVal, err := strconv.Atoi(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(%s): %w", ruleCond, err)
	}

	return func(fieldValue reflect.Value) error {
		v := fieldValue.Int()
		if v < int64(minVal) {
			return ErrIntMin
		}

		return nil
	}, nil
}

// validatorMax is a generator of "max"-rule validator for ruleCond.
// ValueValidatorFn accepts fieldValue of Kind int.
func (r intValidator) validatorMax(ruleCond string) (ValueValidatorFn, error) {
	maxVal, err := strconv.Atoi(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(%s): %w", ruleCond, err)
	}

	return func(fieldValue reflect.Value) error {
		v := fieldValue.Int()
		if int64(maxVal) < v {
			return ErrIntMax
		}

		return nil
	}, nil
}

// validatorIn is a generator of "in"-rule validator for ruleCond.
// ValueValidatorFn accepts fieldValue of Kind int.
func (r intValidator) validatorIn(ruleCond string) (ValueValidatorFn, error) {
	const sep = ","

	inNumbers := strings.Split(ruleCond, sep)
	set := make(map[int64]struct{}, len(inNumbers))

	for pos, ns := range inNumbers {
		n, err := strconv.Atoi(ns)
		if err != nil {
			return nil, fmt.Errorf("strconv.Atoi(%s) at pos %d: %w", ns, pos, err)
		}

		set[int64(n)] = struct{}{}
	}

	return func(fieldValue reflect.Value) error {
		v := fieldValue.Int()
		if _, exists := set[v]; !exists {
			return ErrIntIn
		}

		return nil
	}, nil
}
