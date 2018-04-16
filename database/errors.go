package database

import "fmt"

// EmptyUpdateError represents an error where an update is attempted on the database with no new data.
type EmptyUpdateError struct{}

// NewEmptyUpdateError creates a new empty update error.
func NewEmptyUpdateError() *EmptyUpdateError {
	return &EmptyUpdateError{}
}

// Error makes EmptyUpdateError adhere to the error interface.
func (e *EmptyUpdateError) Error() string {
	return "nothing to update"
}

// NullConstraintError represents an error when an insert or update to the database is attempted with a null field that
// has a non-null constraint.
type NullConstraintError struct {
	field string
}

// NewNullConstraintError creates a new null constraint error with the null field.
func NewNullConstraintError(field string) *NullConstraintError {
	return &NullConstraintError{field: field}
}

// Error makes NullConstraintError adhere to the error interface. The offending column is returned in the string.
func (n *NullConstraintError) Error() string {
	return fmt.Sprintf("%s cannot be null", n.field)
}

// UniqueConstraintError represents an error when an insert or update to the database is attempted that breaks the
// symbol field's unique constraint.
type UniqueConstraintError struct {
	symbol string
}

// NewUniqueConstraintError creates a new unique constraint error.
func NewUniqueConstraintError(symbol string) *UniqueConstraintError {
	return &UniqueConstraintError{symbol: symbol}
}

// Error makes UniqueConstraintError adhere to the error interface.
func (u *UniqueConstraintError) Error() string {
	return fmt.Sprintf("symbol %s already exists", u.symbol)
}

// UnknownIDError represents an error when an update is attempted on a crypto asset with an id that can not be found in
// the database or when the foreign key constraint is broken when trying to insert into the team_member table.
type UnknownIDError struct {
	id int
}

// NewUnknownIDError creates a new unknown ID error with the unknown ID>
func NewUnknownIDError(id int) *UnknownIDError {
	return &UnknownIDError{id: id}
}

// Error makes UnknownIDError adhere to the error interface. The unknown ID is returned in the string.
func (u *UnknownIDError) Error() string {
	return fmt.Sprintf("crypto asset with id %d not found", u.id)
}
