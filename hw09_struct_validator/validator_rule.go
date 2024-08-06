package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrTypeNotSupported             = errors.New("field type is not supported for validation")
	ErrValidatorIncorrectRuleSyntax = errors.New("incorrect rule syntax")
	ErrValidatorRuleNotSupported    = errors.New("rule not supported")
)

type (
	ValueValidatorFn func(fieldValue reflect.Value) error

	// Validator represents validator for some Type.
	Validator interface {
		// String should return validator name.
		String() string

		// Supports indicate if validator supports fieldType for validation.
		Supports(fieldType reflect.Type) bool

		// Kind returns kind of validation.
		Kind() reflect.Kind

		// ValidatorsFor tries to create slice of value validators for provided rules.
		// Returns ErrTypeNotSupported if fieldType is not supported by validator.
		ValidatorsFor(fieldType reflect.Type, rules string) ([]ValueValidatorFn, error)
	}
)

// validatorRuleMatcher is a parser/matcher of validation rules (via validatorsFor).
// Parses rules in format of
//
//	<rule 1><ruleSep><rule 1 condition>[<unionSep><rule 2><ruleSep><rule 2 condition><unionSep>etc...]
type validatorRuleMatcher struct {
	unionSep string
	ruleSep  string
}

// genValidatorFn reoresents generator function for value validator based on ruleCond.
type genValidatorFn func(ruleCond string) (ValueValidatorFn, error)

// validatorsFor returns slice of value validators for provided rules.
// Could return wrapped ErrValidatorIncorrectRuleSyntax or ErrValidatorRuleNotSupported.
func (r validatorRuleMatcher) validatorsFor(rules string, ruleMap map[string]genValidatorFn) ([]ValueValidatorFn, error) {
	vrules := strings.Split(rules, r.unionSep)
	validators := make([]ValueValidatorFn, 0, len(vrules))

	for _, rule := range vrules {
		ruleName, ruleCond, ok := strings.Cut(rule, r.ruleSep)

		// Must be in format "<ruleName><ruleSep><ruleCond>"
		if !ok {
			return nil, fmt.Errorf("%s: %w", rule, ErrValidatorIncorrectRuleSyntax)
		}

		// Check for supported rules
		genFn, exists := ruleMap[ruleName]

		var v ValueValidatorFn
		var err error
		if genFn == nil || !exists {
			// Value validator not found for ruleName
			err = ErrValidatorRuleNotSupported
		} else {
			// Generate value validator of ruleName for ruleCond.
			v, err = genFn(ruleCond)
		}

		// Error while generate value validator.
		if err != nil {
			return nil, fmt.Errorf("%s(%s): %w", ruleName, ruleCond, err)
		}

		validators = append(validators, v)
	}

	return validators, nil
}
