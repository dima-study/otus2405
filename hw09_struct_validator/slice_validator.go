package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// SliceValidator is a Validator for slice of items.
//
// Each item in slice is being validated by suitable validator.
// Argument validators should contains list of supported item validators.
func SliceValidator(validators ...Validator) Validator {
	validator := sliceValidator{
		supported: map[reflect.Kind]Validator{},
	}

	for _, v := range validators {
		validator.supported[v.Kind()] = v
	}

	return validator
}

type sliceValidator struct {
	supported map[reflect.Kind]Validator
}

// SliceItemError represents validation error Err for item in slice at Index position.
type SliceItemError struct {
	Index int
	Err   error
}

func (e SliceItemError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.Index, e.Err)
}

// SliceErrors represents list of validaion errors for items in slice.
type SliceErrors []SliceItemError

func (e SliceErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	errs := strings.Builder{}
	errs.WriteString("slice validation errors:\n")

	for _, ie := range e {
		errs.WriteString("  ")
		errs.WriteString(ie.Error())
		errs.WriteString("\n")
	}

	return errs.String()
}

var ErrSliceNested = errors.New("nested validator error")

// String returns validator name.
func (r sliceValidator) String() string {
	return "sliceValidator"
}

// Supports returns true if fieldType is slice and validator supports
// validation for slice items type.
func (r sliceValidator) Supports(fieldType reflect.Type) bool {
	// Check if structField is type of slice.
	if fieldType.Kind() != reflect.Slice {
		return false
	}

	// Check if slice item type is supported for validation.
	_, exists := r.supported[sliceItemsType(fieldType).Kind()]
	return exists
}

func (r sliceValidator) Kind() reflect.Kind {
	return reflect.Slice
}

// ValidatorsFor returns value validators for provided rules.
// Returns ErrTypeNotSupported if stfieldType is not supported by validator.
func (r sliceValidator) ValidatorsFor(fieldType reflect.Type, rules string) ([]ValueValidatorFn, error) {
	// Check if validator supports specified struct field.
	if !r.Supports(fieldType) {
		return nil, ErrTypeNotSupported
	}

	itemsType := sliceItemsType(fieldType)

	// Suitable validator for structField slice items.
	itemsTypeValidator := r.validatorFor(itemsType)
	// Check if validator exists for provided structField.
	if itemsTypeValidator == nil {
		// Actually must never come here!
		return nil, ErrTypeNotSupported
	}

	// Get itemValidators for slice items.
	itemValidators, err := itemsTypeValidator.ValidatorsFor(itemsType, rules)
	if err != nil {
		return nil, fmt.Errorf("%w for %s: %w", ErrSliceNested, itemsTypeValidator, err)
	}

	return []ValueValidatorFn{r.validatorSlice(itemValidators)}, nil
}

// validatorFor returns suitable validator for slice items type.
func (r sliceValidator) validatorFor(itemsType reflect.Type) Validator {
	validator := r.supported[itemsType.Kind()]
	return validator
}

// validatorSlice is a generator of validator for slice items.
//
// Each slice item will be validated by list of validators.
// ValueValidatorFn returns SliceErrors error or nil.
func (r sliceValidator) validatorSlice(validators []ValueValidatorFn) ValueValidatorFn {
	return func(fieldValue reflect.Value) error {
		errs := SliceErrors{}

		// Each item in slice...
		for i := range fieldValue.Len() {
			itemValue := fieldValue.Index(i)

			// ... should be validated by every validator
			for _, v := range validators {
				// Validate slice item.
				itemErr := v(itemValue)
				if itemErr != nil {
					// Append error once occurred.
					err := SliceItemError{
						Index: i,
						Err:   itemErr,
					}
					errs = append(errs, err)
				}
			}
		}

		if len(errs) != 0 {
			return errs
		}

		return nil
	}
}

// sliceItemsType returns type of items for slice type fieldType.
// fieldType must be type of slice.
func sliceItemsType(fieldType reflect.Type) reflect.Type {
	return fieldType.Elem()
}
