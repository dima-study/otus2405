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
			name:     "stringValidator",
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

// Supports returns true if structField is field of string.
func (r stringValidator) Supports(structField reflect.StructField) bool {
	return structField.Type.Kind() == reflect.String
}

func (r stringValidator) Kind() reflect.Kind {
	return reflect.String
}

// ValidatorsFor returns slice of value validators for provided rules.
func (r stringValidator) ValidatorsFor(rules string) ([]ValueValidatorFn, error) {
	ruleMap := map[string]makeValidatorFn{
		"len":    r.validatorLen,
		"regexp": r.validatorRegexp,
		"in":     r.validatorIn,
	}

	return r.validatorsFor(rules, ruleMap)
}

// validatorLen is generator for "len"-rule validator of ruleCond.
// ValueValidatorFn accepts fieldValue of Kind string.
func (r stringValidator) validatorLen(ruleCond string) (ValueValidatorFn, error) {
	lenVal, err := strconv.Atoi(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("strconv.Atoi(%s): %w", ruleCond, err)
	}

	return func(fieldValue reflect.Value) error {
		s := fieldValue.String()
		if len(s) != lenVal {
			return ErrStringLen
		}

		return nil
	}, nil
}

// validatorRegexp is generator for "regexp"-rule validator of ruleCond.
// ValueValidatorFn accepts fieldValue of Kind string.
func (r stringValidator) validatorRegexp(ruleCond string) (ValueValidatorFn, error) {
	re, err := regexp.Compile(ruleCond)
	if err != nil {
		return nil, fmt.Errorf("regexp.Compile(%s): %w", ruleCond, err)
	}

	return func(fieldValue reflect.Value) error {
		s := fieldValue.String()
		if !re.MatchString(s) {
			return ErrStringRegexp
		}

		return nil
	}, nil
}

// validatorIn is generator for "in"-rule validator of ruleCond.
// ValueValidatorFn accepts fieldValue of Kind string.
func (r stringValidator) validatorIn(ruleCond string) (ValueValidatorFn, error) {
	const sep = ","

	inStrings := strings.Split(ruleCond, sep)
	set := make(map[string]struct{}, len(inStrings))

	for _, s := range inStrings {
		set[s] = struct{}{}
	}

	return func(fieldValue reflect.Value) error {
		v := fieldValue.String()
		if _, exists := set[v]; !exists {
			return ErrStringIn
		}

		return nil
	}, nil
}
