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

var ErrSliceNested = errors.New("nested validator error")

// String returns validator name.
func (r sliceValidator) String() string {
	return "sliceValidator"
}

// Supports returns true if t is type of slice and validator supports
// validation for slice items type.
func (r sliceValidator) Supports(t reflect.Type) bool {
	// Check if t is type of slice.
	if t.Kind() != reflect.Slice {
		return false
	}

	// Check if slice item type is supported for validation.
	_, exists := r.supported[t.Elem().Kind()]
	return exists
}

func (r sliceValidator) Kind() reflect.Kind {
	return reflect.Slice
}

// ValidatorsFor returns value validators for provided rules.
//
// Returns ErrTypeNotSupported if sliceType is not supported by validator.
// Could return (possibly wrapped) ErrSliceNested.
func (r sliceValidator) ValidatorsFor(sliceType reflect.Type, rules []Rule) ([]ValueValidatorFn, error) {
	// Check if validator supports provided type.
	if !r.Supports(sliceType) {
		return nil, ErrTypeNotSupported
	}

	itemsType := sliceType.Elem()

	// Suitable validator for slice items.
	itemsTypeValidator := r.validatorFor(itemsType)

	// Check if validator exists for provided items type.
	if itemsTypeValidator == nil {
		// Must NEVER come here!
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

// SliceItemError represents validation error Err for item in slice at Index position.
type SliceItemError struct {
	Parent string
	Index  int
	Err    error
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

// validatorSlice is a generator of validator for slice items.
//
// Each slice item will be validated by list of validators.
// ValueValidatorFn accepts value of Kind slice, and returns SliceErrors error or nil.
func (r sliceValidator) validatorSlice(validators []ValueValidatorFn) ValueValidatorFn {
	return func(sliceValue reflect.Value) error {
		errs := SliceErrors{}

		// Each item in slice...
		for i := range sliceValue.Len() {
			itemValue := sliceValue.Index(i)

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
