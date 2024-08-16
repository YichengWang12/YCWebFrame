package errs

import (
	"errors"
	"fmt"
)

var (
	// ErrPointerOnly Only supports one-level pointer as input
	// Seeing this error means you input something else
	// We don't want users to be able to use err == ErrPointerOnly directly
	// So put it in our internal package
	ErrPointerOnly = errors.New("orm: Only supports one-level pointer as input, such as *User")
	// ErrNoRows represents no data found
	ErrNoRows                 = errors.New("orm: no data found")
	ErrTooManyReturnedColumns = errors.New("orm: too many columns")
)

// NewErrUnknownField returns an error representing an unknown field
// Generally means you may have entered a column name, or entered the wrong field name
func NewErrUnknownField(fd string) error {
	return fmt.Errorf("orm: unknown field %s", fd)
}

// NewErrUnsupportedExpressionType returns an error message that does not support the expression
func NewErrUnsupportedExpressionType(exp any) error {
	return fmt.Errorf("orm: unsupported expression: %v ", exp)
}

func NewErrInvalidTagContent(tag string) error {
	return fmt.Errorf("orm: invalid tag content %s", tag)
}

func NewErrUnknownColumn(col string) error {
	return fmt.Errorf("orm: unknown column: %s", col)
}
