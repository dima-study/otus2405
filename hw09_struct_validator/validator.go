package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
)

// DefaultValidators contains validators supported by Validate.
// Validators could be added or removed on depend.
//
// By default next validators are supported:
//   - IntValidator
//   - StringValidator
//   - SliceValidator for ints and strings
var DefaultValidators = []Validator{
	IntValidator(),
	StringValidator(),
	SliceValidator(IntValidator(), StringValidator()),
}

// ValidationError is an alias to FieldError of StructValidator.
type ValidationError = FieldError

// ValidationErrors is an alias to StructErrors of StructValidator.
type ValidationErrors = StructErrors

// Validate tries to validate fields for v, where v is type of struct.
func Validate(v any) error {
	// Create sruct validator with default validators.
	validator := StructValidator(DefaultValidators...)

	vType := reflect.TypeOf(v)

	// Get struct validators
	validators, err := validator.ValidatorsFor(vType, []Rule{{Name: RuleStructNested}})
	if err != nil {
		return fmt.Errorf("struct validator: %w", err)
	}

	vValue := reflect.ValueOf(v)

	err = validators[0](vValue)
	if err != nil {
		var structErr ValidationErrors
		switch {
		case errors.As(err, &structErr):
			return structErr
		default:
			return err
		}
	}

	// Success validation.
	return nil
}
