package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// StructValidator is a Validator for struct fields.
//
// Struct fields are being validated by its suitable validators based on "validate" field tag,
// when it is provided.
//
// "validate" field tag is being used to specify validation rules for the public fields.
// Supports union (|) of rules.
//
// Argument validators should contains list of supported validators.
func StructValidator(validators ...Validator) Validator {
	validator := structValidator{
		supported: map[reflect.Kind]Validator{},
	}

	for _, v := range validators {
		validator.supported[v.Kind()] = v
	}

	return validator
}

// # Supported validation rules.
//
// RuleStructNested indicates that field value (which is type of struct) should be validated.
const RuleStructNested = "nested"

type structValidator struct {
	supported map[reflect.Kind]Validator
}

var ErrStructNested = errors.New("nested validator error")

// String returns validator name.
func (r structValidator) String() string {
	return "structValidator"
}

// Supports returns true if t is type of struct.
func (r structValidator) Supports(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func (r structValidator) Kind() reflect.Kind {
	return reflect.Struct
}

// ValidatorsFor returns value validators for provided rules.
//
// Returns ErrTypeNotSupported if structType is not supported by validator.
// Could return (possibly wrapped) ErrStructNested or ErrRuleNotSupported.
func (r structValidator) ValidatorsFor(structType reflect.Type, rules []Rule) ([]ValueValidatorFn, error) {
	// Check if validator supports provided type.
	if !r.Supports(structType) {
		return nil, ErrTypeNotSupported
	}

	valueValidators := make([]ValueValidatorFn, 0, len(rules))

	for _, rule := range rules {
		switch rule.Name {
		case RuleStructNested:
			valueValidator, err := r.validatorNested(structType)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrStructNested, err)
			}

			valueValidators = append(valueValidators, valueValidator)
		default:
			return nil, fmt.Errorf("%s(%s): %w", rule.Name, rule.Condition, ErrRuleNotSupported)
		}
	}

	return valueValidators, nil
}

// validatorFor returns suitable validator for struct items type.
func (r structValidator) validatorFor(supportedType reflect.Type) Validator {
	validator, exists := r.supported[supportedType.Kind()]

	// If struct validator not provided,
	// then return current struct validator to validate structs.
	if !exists && supportedType.Kind() == reflect.Struct {
		return r
	}

	return validator
}

type fieldValidators struct {
	Field      reflect.StructField
	Validators []ValueValidatorFn
}

// FieldError is an error of validation for specific field.
type FieldError struct {
	Field string
	Err   error
}

func (e FieldError) Error() string {
	return fmt.Sprintf("{%s}: %s", e.Field, e.Err)
}

// StructErrors contains list of validaion errors of struct fields.
type StructErrors []FieldError

func (e StructErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	errs := strings.Builder{}
	errs.WriteString("struct validation errors:\n")

	for i, ie := range e {
		if i > 0 {
			errs.WriteString("\n")
		}

		e := strings.ReplaceAll(ie.Error(), "\n", "\n  ")
		errs.WriteString("  ")
		errs.WriteString(e)
	}

	return errs.String()
}

// validatorNested is a generator of "nested"-rule validator for struct.
// ValueValidatorFn accepts value of Kind struct, and returns StructErrors or nil.
func (r structValidator) validatorNested(structType reflect.Type) (ValueValidatorFn, error) {
	fieldValidatorsList, err := r.getFieldsValidators(structType)
	if err != nil {
		return nil, err
	}

	return func(structValue reflect.Value) error {
		errs := StructErrors{}

		// For each field validators list...
		for _, v := range fieldValidatorsList {
			fieldValue := structValue.FieldByName(v.Field.Name)

			// ... validate field value against each validator.
			for _, valueValidator := range v.Validators {
				valueErr := valueValidator(fieldValue)

				// Collect error if occurred.
				if valueErr != nil {
					err := FieldError{
						Field: v.Field.Name,
						Err:   valueErr,
					}
					errs = append(errs, err)
				}
			}
		}

		if len(errs) != 0 {
			return errs
		}

		return nil
	}, nil
}

// getFieldsValidators returns list of fields with "validate" tag and its validators.
//
// Returns the first occurred error.
func (r structValidator) getFieldsValidators(structType reflect.Type) ([]fieldValidators, error) {
	const tagName = "validate"

	// Each element of validators contains validators for corresponded field.
	validators := make([]fieldValidators, 0, structType.NumField())

	// For each field of struct...
	for i := range structType.NumField() {
		field := structType.Field(i)

		// Skip non-exported fields for validation.
		if !field.IsExported() {
			continue
		}

		// Validate field only if struct tag is provided.
		rulesTag, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		// ... try to get type validators
		typeValidator := r.validatorFor(field.Type)
		if typeValidator == nil {
			return nil, fmt.Errorf("field %s of type %s: %w", field.Name, field.Type.Kind(), ErrTypeNotSupported)
		}

		// Parse validation rules
		rules := parseRules(rulesTag)

		// Get type validators for specified field rules
		fValidators, err := typeValidator.ValidatorsFor(field.Type, rules)
		if err != nil {
			return nil, fmt.Errorf("field %s of type %s: %w", field.Name, field.Type.Kind(), err)
		}

		// Field validators could be empty: skip for validation
		if len(fValidators) == 0 {
			continue
		}

		v := fieldValidators{
			Field:      field,
			Validators: fValidators,
		}

		validators = append(validators, v)
	}

	return validators, nil
}

// parseRules tries to parse rulesTag (from "validation" field tag) into slice of Rules.
// Parses rules in format of:
//
//	<rule 1 name>[:<rule 1 condition>[|<rule 2 name>:<rule 2 condition>|etc...]]
func parseRules(rulesTag string) []Rule {
	rulesList := strings.Split(rulesTag, "|")
	rules := make([]Rule, 0, len(rulesList))

	// TODO: unique rules
	for _, rule := range rulesList {
		name, condition, ok := strings.Cut(rule, ":")

		// Must be in format "<name>:<condition>" or "<name>"
		if ok {
			rules = append(rules, Rule{name, condition})
		} else {
			rules = append(rules, Rule{name, ""})
		}
	}

	return rules
}
