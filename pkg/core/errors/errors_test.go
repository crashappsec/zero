package errors

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	// Verify all sentinel errors are distinct
	sentinels := []error{
		ErrNotFound,
		ErrInvalid,
		ErrTimeout,
		ErrUnauthorized,
		ErrForbidden,
		ErrRateLimited,
		ErrUnavailable,
		ErrCancelled,
		ErrAlreadyExists,
		ErrDependency,
		ErrConfiguration,
		ErrToolNotFound,
	}

	for i, err1 := range sentinels {
		if err1 == nil {
			t.Errorf("Sentinel error %d is nil", i)
		}
		for j, err2 := range sentinels {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("Sentinel errors %d and %d should be distinct", i, j)
			}
		}
	}
}

func TestWrap(t *testing.T) {
	base := errors.New("base error")
	wrapped := Wrap(base, "context")

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}
	if !errors.Is(wrapped, base) {
		t.Error("Wrapped error should contain base error")
	}

	// Wrap nil should return nil
	if Wrap(nil, "context") != nil {
		t.Error("Wrap(nil, ...) should return nil")
	}
}

func TestWrapf(t *testing.T) {
	base := errors.New("base error")
	wrapped := Wrapf(base, "operation %s failed", "test")

	if wrapped == nil {
		t.Fatal("Wrapf returned nil")
	}
	if !errors.Is(wrapped, base) {
		t.Error("Wrapped error should contain base error")
	}

	// Wrapf nil should return nil
	if Wrapf(nil, "context") != nil {
		t.Error("Wrapf(nil, ...) should return nil")
	}
}

func TestIs(t *testing.T) {
	wrapped := Wrap(ErrNotFound, "resource")
	if !Is(wrapped, ErrNotFound) {
		t.Error("Is should return true for wrapped error")
	}
	if Is(wrapped, ErrTimeout) {
		t.Error("Is should return false for different error")
	}
}

func TestNotFoundError(t *testing.T) {
	err := NotFoundError("config file")
	if !IsNotFound(err) {
		t.Error("NotFoundError should be IsNotFound")
	}
}

func TestInvalidError(t *testing.T) {
	err := InvalidError("input parameter")
	if !IsInvalid(err) {
		t.Error("InvalidError should be IsInvalid")
	}
}

func TestTimeoutError(t *testing.T) {
	err := TimeoutError("API call")
	if !IsTimeout(err) {
		t.Error("TimeoutError should be IsTimeout")
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := UnauthorizedError("missing token")
	if !IsUnauthorized(err) {
		t.Error("UnauthorizedError should be IsUnauthorized")
	}
}

func TestToolNotFoundError(t *testing.T) {
	err := ToolNotFoundError("semgrep")
	if !IsToolNotFound(err) {
		t.Error("ToolNotFoundError should be IsToolNotFound")
	}
}

func TestDependencyError(t *testing.T) {
	base := errors.New("connection failed")
	err := DependencyError("database", base)
	if !Is(err, ErrDependency) {
		t.Error("DependencyError should wrap ErrDependency")
	}

	// Without underlying error
	err2 := DependencyError("redis", nil)
	if !Is(err2, ErrDependency) {
		t.Error("DependencyError without cause should still be ErrDependency")
	}
}

func TestConfigError(t *testing.T) {
	base := errors.New("parse failed")
	err := ConfigError("zero.config.json", base)
	if !Is(err, ErrConfiguration) {
		t.Error("ConfigError should wrap ErrConfiguration")
	}
}

func TestMultiError(t *testing.T) {
	multi := NewMultiError()

	if multi.HasErrors() {
		t.Error("New MultiError should have no errors")
	}

	if multi.ErrorOrNil() != nil {
		t.Error("ErrorOrNil should return nil when empty")
	}

	multi.Add(errors.New("error 1"))
	multi.Add(nil) // Should be ignored
	multi.Add(errors.New("error 2"))

	if !multi.HasErrors() {
		t.Error("MultiError should have errors after Add")
	}

	if len(multi.Errors) != 2 {
		t.Errorf("MultiError should have 2 errors, got %d", len(multi.Errors))
	}

	if multi.ErrorOrNil() == nil {
		t.Error("ErrorOrNil should return error when not empty")
	}

	errStr := multi.Error()
	if errStr == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestMultiError_SingleError(t *testing.T) {
	multi := NewMultiError()
	base := errors.New("single error")
	multi.Add(base)

	if multi.Error() != base.Error() {
		t.Errorf("Single error MultiError should return that error's message")
	}
}

func TestMultiError_Empty(t *testing.T) {
	multi := NewMultiError()
	if multi.Error() != "no errors" {
		t.Errorf("Empty MultiError.Error() = %q, want %q", multi.Error(), "no errors")
	}
}

func TestNew(t *testing.T) {
	err := New("test error")
	if err == nil {
		t.Error("New should return non-nil error")
	}
	if err.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error")
	}
}

func TestNewf(t *testing.T) {
	err := Newf("error %d: %s", 42, "test")
	if err == nil {
		t.Error("Newf should return non-nil error")
	}
	expected := "error 42: test"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestJoin(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	joined := Join(err1, err2)
	if joined == nil {
		t.Error("Join should return non-nil error")
	}

	if !errors.Is(joined, err1) {
		t.Error("Joined error should contain err1")
	}
	if !errors.Is(joined, err2) {
		t.Error("Joined error should contain err2")
	}
}
