package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// StructValidator is a Validator for struct fields.
//
// Struct fields are being validated by its suitable validator based on "validate" field tag,
// when it is provided and not empty.
// "validate" field tag is being used to specify validation rules for the field.
//
// Supported validation rules:
//
//	nested - indicates that field value (which is type of struct) should be validated.
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

type structValidator struct {
	supported map[reflect.Kind]Validator
}

var ErrStructNested = errors.New("nested validator error")

// String returns validator name.
func (r structValidator) String() string {
	return "structValidator"
}

// Supports returns true if fieldType is struct.
func (r structValidator) Supports(fieldType reflect.Type) bool {
	return fieldType.Kind() == reflect.Struct
}

func (r structValidator) Kind() reflect.Kind {
	return reflect.Struct
}

const RuleStructNested = "nested"

// ValidatorsFor returns value validators for provided rules.
//
// Returns ErrTypeNotSupported if stfieldType is not supported by validator.
// Could return (possibly wrapped) ErrStructNested or ErrValidatorRuleNotSupported.
func (r structValidator) ValidatorsFor(fieldType reflect.Type, rules string) ([]ValueValidatorFn, error) {
	// Check if validator supports specified struct field.
	if !r.Supports(fieldType) {
		return nil, ErrTypeNotSupported
	}

	switch rules {
	case RuleStructNested:
		fieldsValidator, err := r.validatorNested(fieldType)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrStructNested, err)
		}
		return []ValueValidatorFn{fieldsValidator}, nil
	case "":
		// Rule is empty - no validation needed
		return nil, nil
	default:
		return nil, ErrValidatorRuleNotSupported
	}
}

// validatorFor returns suitable validator for struct items type.
func (r structValidator) validatorFor(supportedType reflect.Type) Validator {
	validator := r.supported[supportedType.Kind()]
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
		tagRule, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		// ... try to get type validators
		typeValidator := r.validatorFor(field.Type)
		if typeValidator == nil {
			return nil, fmt.Errorf("field %s of type %s: %w", field.Name, field.Type.Kind(), ErrTypeNotSupported)
		}

		// Get type validators for specified field rules
		fValidators, err := typeValidator.ValidatorsFor(field.Type, tagRule)
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
