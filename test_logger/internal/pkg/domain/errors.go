package domain

import (
	"fmt"
)

// ValidationError represents an error that occurs when data fails validation
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

// BadRequestError represents an error that occurs when a request is malformed
type BadRequestError struct {
	Message string
	Err     error
}

func (e *BadRequestError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string, err error) error {
	return &BadRequestError{
		Message: message,
		Err:     err,
	}
}

// NotFoundError represents an error that occurs when an entity is not found
type NotFoundError struct {
	EntityType string
	ID         string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.EntityType, e.ID)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(entityType, id string) error {
	return &NotFoundError{
		EntityType: entityType,
		ID:         id,
	}
}

// ConflictError represents an error that occurs when there's a conflict
// (e.g., duplicate entry, concurrency issue)
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) error {
	return &ConflictError{Message: message}
}

// UnauthorizedError represents an error that occurs when a user is not authorized
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) error {
	return &UnauthorizedError{Message: message}
}

// ForbiddenError represents an error that occurs when an operation is forbidden
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) error {
	return &ForbiddenError{Message: message}
}

// InternalError represents an internal server error
type InternalError struct {
	Message string
	Err     error
}

func (e *InternalError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) error {
	return &InternalError{
		Message: message,
		Err:     err,
	}
} 