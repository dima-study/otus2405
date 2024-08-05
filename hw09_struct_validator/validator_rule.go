package hw09structvalidator

import (
	"fmt"
	"reflect"
	"strings"
	"errors"
)

var (
	ErrValidatorIncorrectRuleSyntax = errors.New("incorrect rule syntax")
	ErrValidatorRuleNotSupported    = errors.New("rule not supported")
)

type (
	ValueValidatorFn func(fieldValue reflect.Value) error

	// Validator represents validator for some Kind of struct field.
	Validator interface {
		// Supports indicate if rule supports validation for structField
		Supports(structField reflect.StructField) bool

		// Kind returns kind of validation.
		Kind() reflect.Kind

		// ValidatorsFor tries to create slice of value validators for provided rules.
		ValidatorsFor(rules string) ([]ValueValidatorFn, error)
	}
)

type validatorRuleMatcher struct {
	name     string
	unionSep string
	ruleSep  string
}

type makeValidatorFn func(string) (ValueValidatorFn, error)

// validatorsFor returns slice of value validators for provided rules.
// Could return wrapped ErrValidatorIncorrectRuleSyntax or ErrValidatorRuleNotSupported.
func (r validatorRuleMatcher) validatorsFor(rules string, ruleMap map[string]makeValidatorFn) ([]ValueValidatorFn, error) {
	vrules := strings.Split(rules, r.unionSep)
	validators := make([]ValueValidatorFn, 0, len(vrules))

	for _, rule := range vrules {
		ruleName, ruleCond, ok := strings.Cut(rule, r.ruleSep)

		// Must be in format "<ruleName><ruleSep><ruleCond>"
		if !ok {
			return nil, fmt.Errorf("%s %s: %w", r.name, rule, ErrValidatorIncorrectRuleSyntax)
		}

		// Check for supported rules
		fn, exists := ruleMap[ruleName]

		var v ValueValidatorFn
		var err error
		if fn == nil || !exists {
			err = ErrValidatorRuleNotSupported
		} else {
			v, err = fn(ruleCond)
		}

		// Validator preparing error.
		if err != nil {
			return nil, fmt.Errorf("%s %s(%s): %w", r.name, ruleName, ruleCond, err)
		}

		validators = append(validators, v)
	}

	return validators, nil
}
