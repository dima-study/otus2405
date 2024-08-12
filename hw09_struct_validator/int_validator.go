package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// IntValidator is a Validator with validation rules for int type.
func IntValidator() Validator {
	return intValidator{}
}

// # Supported validation rules.
const (
	// RuleIntMin validates value must not be less than the number.
	//
	//	min:number
	RuleIntMin = "min"

	// RuleIntMax validates value must not be great than the number.
	//
	//	max:number
	RuleIntMax = "max"

	// RuleIntIn validates value must be in the set of numbers {n1,n2,...,nN}
	//
	//	in:n1,n2,...,nN
	RuleIntIn = "in"
)

type intValidator struct{}

var (
	ErrIntMin = errors.New("value less than min")
	ErrIntMax = errors.New("value great than max")
	ErrIntIn  = errors.New("value not in the set")
)

// String returns validator name.
func (r intValidator) String() string {
	return "intValidator"
}

// Supports returns true if t is type of int.
func (r intValidator) Supports(t reflect.Type) bool {
	return t.Kind() == reflect.Int
}

func (r intValidator) Kind() reflect.Kind {
	return reflect.Int
}

// ValidatorsFor returns slice of value validators for provided rules.
//
// Returns ErrTypeNotSupported if intType is not supported by validator.
func (r intValidator) ValidatorsFor(intType reflect.Type, rules []Rule) ([]ValueValidatorFn, error) {
	// Check if validator supports provided type.
	if !r.Supports(intType) {
		return nil, ErrTypeNotSupported
	}

	valueValidators := make([]ValueValidatorFn, 0, len(rules))

	// Map supported rules to its validator generators.
	ruleMap := map[string]func(ruleCond string) (ValueValidatorFn, error){
		RuleIntMin: r.validatorMin,
		RuleIntMax: r.validatorMax,
		RuleIntIn:  r.validatorIn,
	}

	// For each rule...
	for _, rule := range rules {
		// ...check if rule is supported
		fn, exists := ruleMap[rule.Name]
		if fn == nil || !exists {
			// Value validator not found for the rule
			return nil, fmt.Errorf("%s(%s): %w", rule.Name, rule.Condition, ErrRuleNotSupported)
		}

		// Generate value validator for the rule
		valueValidator, err := fn(rule.Condition)
		if err != nil {
			// Error while generate value validator.
			return nil, fmt.Errorf("%s(%s): %w: %w", rule.Name, rule.Condition, ErrRuleInvalidCondition, err)
		}

		valueValidators = append(valueValidators, valueValidator)
	}

	return valueValidators, nil
}

// validatorMin is a generator of "min"-rule validator for ruleCond.
// ValueValidatorFn accepts value of Kind int, and returns ErrIntMin or nil.
func (r intValidator) validatorMin(ruleCond string) (ValueValidatorFn, error) {
	minVal, err := strconv.Atoi(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(%s): %w", ruleCond, err)
	}

	return func(intValue reflect.Value) error {
		v := intValue.Int()
		if v < int64(minVal) {
			return ErrIntMin
		}

		return nil
	}, nil
}

// validatorMax is a generator of "max"-rule validator for ruleCond.
// ValueValidatorFn accepts value of Kind int, and returns ErrIntMax or nil.
func (r intValidator) validatorMax(ruleCond string) (ValueValidatorFn, error) {
	maxVal, err := strconv.Atoi(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(%s): %w", ruleCond, err)
	}

	return func(intValue reflect.Value) error {
		v := intValue.Int()
		if int64(maxVal) < v {
			return ErrIntMax
		}

		return nil
	}, nil
}

// validatorIn is a generator of "in"-rule validator for ruleCond.
// ValueValidatorFn accepts value of Kind int, and returns ErrIntIn or nil.
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

	return func(intValue reflect.Value) error {
		v := intValue.Int()
		if _, exists := set[v]; !exists {
			return ErrIntIn
		}

		return nil
	}, nil
}
