package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

var ErrNotStruct = errors.New("v must be type of struct")

func (v ValidationErrors) Error() string {
	panic("implement me")
}

func Validate(v interface{}) error {
	vType := reflect.TypeOf(v)
	if vType.Kind() != reflect.Struct {
		return fmt.Errorf("v type Kind is %s, %w", vType.Kind(), ErrNotStruct)
	}
	// Place your code here.
	return nil
}
