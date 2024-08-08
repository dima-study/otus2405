package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// StringValidator is a Validator with validation rules for string type.
//
// Supported validation rules:
//
//	len:number      - string length must be equal to the number
//	regexp:re       - string must match the regular expression re
//	in:s1,s2,...,sN - string must be in the set of strings {s1,s2,...,sN}
//
// Supports union (|) of rules.
// Example: "len:42|re:^\\d+$" - string length must be 42 and contain only didgits.
func StringValidator() Validator {
	return stringValidator{
		validatorRuleMatcher: validatorRuleMatcher{
			unionSep: "|",
			ruleSep:  ":",
		},
	}
}

type stringValidator struct {
	validatorRuleMatcher
}

var (
	ErrStringLen    = errors.New("string length not equal to len")
	ErrStringRegexp = errors.New("string does not match regexp")
	ErrStringIn     = errors.New("string not in the set")
)

// String returns validator name.
func (r stringValidator) String() string {
	return "stringValidator"
}

// Supports returns true if fieldType is string.
func (r stringValidator) Supports(fieldType reflect.Type) bool {
	return fieldType.Kind() == reflect.String
}

func (r stringValidator) Kind() reflect.Kind {
	return reflect.String
}

const (
	RuleStringLen    = "len"
	RuleStringRegexp = "regexp"
	RuleStringIn     = "in"
)

// ValidatorsFor returns slice of value validators for provided rules.
//
// Returns ErrTypeNotSupported if fieldType is not supported by validator.
func (r stringValidator) ValidatorsFor(fieldType reflect.Type, rules string) ([]ValueValidatorFn, error) {
	// Check if validator supports specified struct field.
	if !r.Supports(fieldType) {
		return nil, ErrTypeNotSupported
	}

	ruleMap := map[string]genValidatorFn{
		RuleStringLen:    r.validatorLen,
		RuleStringRegexp: r.validatorRegexp,
		RuleStringIn:     r.validatorIn,
	}

	return r.matchedValidatorsFor(rules, ruleMap)
}

// validatorLen is a generator of "len"-rule validator for ruleCond.
// ValueValidatorFn accepts value of Kind string, and returns ErrStringLen or nil.
func (r stringValidator) validatorLen(ruleCond string) (ValueValidatorFn, error) {
	lenVal, err := strconv.Atoi(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(%s): %w", ruleCond, err)
	}

	return func(stringValue reflect.Value) error {
		s := stringValue.String()
		if len(s) != lenVal {
			return ErrStringLen
		}

		return nil
	}, nil
}

// validatorRegexp is a generator of "regexp"-rule validator for ruleCond.
// ValueValidatorFn accepts value of Kind string, and returns ErrStringRegexp or nil.
func (r stringValidator) validatorRegexp(ruleCond string) (ValueValidatorFn, error) {
	re, err := regexp.Compile(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("regexp.Compile(%s): %w", ruleCond, err)
	}

	return func(stringValue reflect.Value) error {
		s := stringValue.String()
		if !re.MatchString(s) {
			return ErrStringRegexp
		}

		return nil
	}, nil
}

// validatorIn is a generator of "in"-rule validator for ruleCond.
// ValueValidatorFn accepts value of Kind string, and returns ErrStringIn or nil.
func (r stringValidator) validatorIn(ruleCond string) (ValueValidatorFn, error) {
	const sep = ","

	inStrings := strings.Split(ruleCond, sep)
	set := make(map[string]struct{}, len(inStrings))

	for _, s := range inStrings {
		set[s] = struct{}{}
	}

	return func(stringValue reflect.Value) error {
		v := stringValue.String()
		if _, exists := set[v]; !exists {
			return ErrStringIn
		}

		return nil
	}, nil
}
