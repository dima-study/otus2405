package hw09structvalidator

import (
	"errors"
	"reflect"
)

var (
	ErrTypeNotSupported = errors.New("type is not supported for validation")

	ErrRuleIncorrectSyntax  = errors.New("incorrect rule syntax")
	ErrRuleInvalidCondition = errors.New("invalid rule condition")
	ErrRuleNotSupported     = errors.New("rule not supported")
)

type (
	// ValueValidatorFn represents validator for the value.
	// Returns error on failed validation.
	ValueValidatorFn func(value reflect.Value) error

	// Rule represents validation rule.
	Rule struct {
		Name      string
		Condition string
	}

	// Validator represents validator for some Type.
	Validator interface {
		// String should return validator name.
		String() string

		// Supports indicate if validator supports type t for validation.
		Supports(t reflect.Type) bool

		// Kind returns kind of validator.
		Kind() reflect.Kind

		// ValidatorsFor returns value validators for provided rules.
		// Returns ErrTypeNotSupported (possibly wrapped) if fieldType is not supported by validator.
		ValidatorsFor(fieldType reflect.Type, rules []Rule) ([]ValueValidatorFn, error)
	}
)
