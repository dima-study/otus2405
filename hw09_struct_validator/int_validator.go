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
			name:     "intValidator",
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

// Supports returns true if structField is field of int.
func (r intValidator) Supports(structField reflect.StructField) bool {
	return structField.Type.Kind() == reflect.Int
}

func (r intValidator) Kind() reflect.Kind {
	return reflect.Int
}

// ValidatorsFor returns slice of value validators for provided rules.
func (r intValidator) ValidatorsFor(rules string) ([]ValueValidatorFn, error) {
	ruleMap := map[string]makeValidatorFn{
		"min": r.validatorMin,
		"max": r.validatorMax,
		"in":  r.validatorIn,
	}

	return r.validatorsFor(rules, ruleMap)
}

// validatorMin is generator for "min"-rule validator of ruleCond.
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

// validatorMax is generator for "max"-rule validator of ruleCond.
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

// validatorIn is generator for "in"-rule validator of ruleCond.
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
