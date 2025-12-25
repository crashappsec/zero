// Package errors provides sentinel errors and error utilities for Zero.
// Use these errors for consistent error handling across the codebase.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure scenarios
var (
	// ErrNotFound indicates a requested resource was not found
	ErrNotFound = errors.New("not found")

	// ErrInvalid indicates invalid input or configuration
	ErrInvalid = errors.New("invalid")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("timeout")

	// ErrUnauthorized indicates missing or invalid authentication
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the operation is not permitted
	ErrForbidden = errors.New("forbidden")

	// ErrRateLimited indicates too many requests
	ErrRateLimited = errors.New("rate limited")

	// ErrUnavailable indicates a service or resource is unavailable
	ErrUnavailable = errors.New("unavailable")

	// ErrCancelled indicates the operation was cancelled
	ErrCancelled = errors.New("cancelled")

	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errors.New("already exists")

	// ErrDependency indicates a dependency is missing or failed
	ErrDependency = errors.New("dependency error")

	// ErrConfiguration indicates a configuration problem
	ErrConfiguration = errors.New("configuration error")

	// ErrToolNotFound indicates an external tool is not installed
	ErrToolNotFound = errors.New("tool not found")
)

// Is reports whether any error in err's tree matches target.
// This is a convenience wrapper around errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's tree that matches target.
// This is a convenience wrapper around errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with additional context.
// Returns nil if err is nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message.
// Returns nil if err is nil.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// New creates a new error with the given message.
// This is a convenience wrapper around errors.New.
func New(message string) error {
	return errors.New(message)
}

// Newf creates a new error with a formatted message.
func Newf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// NotFoundError creates a not found error with context
func NotFoundError(resource string) error {
	return fmt.Errorf("%s: %w", resource, ErrNotFound)
}

// InvalidError creates an invalid input error with context
func InvalidError(what string) error {
	return fmt.Errorf("%s: %w", what, ErrInvalid)
}

// TimeoutError creates a timeout error with context
func TimeoutError(operation string) error {
	return fmt.Errorf("%s: %w", operation, ErrTimeout)
}

// UnauthorizedError creates an unauthorized error with context
func UnauthorizedError(reason string) error {
	return fmt.Errorf("%s: %w", reason, ErrUnauthorized)
}

// ToolNotFoundError creates a tool not found error
func ToolNotFoundError(tool string) error {
	return fmt.Errorf("%s: %w", tool, ErrToolNotFound)
}

// DependencyError creates a dependency error with context
func DependencyError(dep string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w: %v", dep, ErrDependency, err)
	}
	return fmt.Errorf("%s: %w", dep, ErrDependency)
}

// ConfigError creates a configuration error with context
func ConfigError(what string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w: %v", what, ErrConfiguration, err)
	}
	return fmt.Errorf("%s: %w", what, ErrConfiguration)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalid checks if an error is an invalid error
func IsInvalid(err error) bool {
	return errors.Is(err, ErrInvalid)
}

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsRateLimited checks if an error is a rate limited error
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsCancelled checks if an error is a cancelled error
func IsCancelled(err error) bool {
	return errors.Is(err, ErrCancelled)
}

// IsToolNotFound checks if an error is a tool not found error
func IsToolNotFound(err error) bool {
	return errors.Is(err, ErrToolNotFound)
}

// Join combines multiple errors into a single error.
// This is a convenience wrapper around errors.Join (Go 1.20+).
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

// Error implements the error interface
func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}
	return fmt.Sprintf("%d errors: %v", len(m.Errors), m.Errors[0])
}

// Add adds an error to the multi-error
func (m *MultiError) Add(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (m *MultiError) HasErrors() bool {
	return len(m.Errors) > 0
}

// ErrorOrNil returns nil if there are no errors, otherwise returns the MultiError
func (m *MultiError) ErrorOrNil() error {
	if !m.HasErrors() {
		return nil
	}
	return m
}

// NewMultiError creates a new MultiError
func NewMultiError() *MultiError {
	return &MultiError{}
}
