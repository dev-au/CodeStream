package customerrors_test

import (
	"errors"
	"fmt"
	"testing"

	customErrors "github.com/dev-au/CodeStream/pkg/customerrors"
)

func TestWrapSystemError(t *testing.T) {
	inner := fmt.Errorf("disk full")
	wrapped := customErrors.WrapSystemError(inner)

	if !errors.Is(wrapped, customErrors.ErrSystem) {
		t.Error("expected wrapped error to match ErrSystem")
	}
}

func TestWrapExternalServiceError(t *testing.T) {
	inner := fmt.Errorf("service unreachable")
	wrapped := customErrors.WrapExternalServiceError(inner)

	if !errors.Is(wrapped, customErrors.ErrExternalService) {
		t.Error("expected wrapped error to match ErrExternalService")
	}
}

func TestWrapDataNotFoundError(t *testing.T) {
	inner := fmt.Errorf("row not found")
	wrapped := customErrors.WrapDataNotFoundError(inner)

	if !errors.Is(wrapped, customErrors.ErrDataNotFound) {
		t.Error("expected wrapped error to match ErrDataNotFound")
	}
}

func TestWrapValidationError(t *testing.T) {
	inner := fmt.Errorf("field required")
	wrapped := customErrors.WrapValidationError(inner)

	if !errors.Is(wrapped, customErrors.ErrValidation) {
		t.Error("expected wrapped error to match ErrValidation")
	}
}

func TestWrapPermissionDeniedError(t *testing.T) {
	inner := fmt.Errorf("access denied")
	wrapped := customErrors.WrapPermissionDeniedError(inner)

	if !errors.Is(wrapped, customErrors.ErrPermissionDenied) {
		t.Error("expected wrapped error to match ErrPermissionDenied")
	}
}

func TestCustomError_Error_ReturnsInnerMessage(t *testing.T) {
	inner := fmt.Errorf("the real error message")
	wrapped := customErrors.WrapSystemError(inner)

	if wrapped.Error() != inner.Error() {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), inner.Error())
	}
}

func TestCustomError_Unwrap_ExposesInner(t *testing.T) {
	inner := fmt.Errorf("inner cause")
	wrapped := customErrors.WrapSystemError(inner)

	if !errors.Is(wrapped, inner) {
		t.Error("errors.Is should find the inner error via Unwrap")
	}
}

func TestCustomError_Is_SameBase_Matches(t *testing.T) {
	err1 := customErrors.WrapSystemError(fmt.Errorf("a"))
	err2 := customErrors.WrapSystemError(fmt.Errorf("b"))

	if !errors.Is(err1, err2) {
		t.Error("two errors with the same base type should match via errors.Is")
	}
}

func TestCustomError_Is_DifferentBase_DoesNotMatch(t *testing.T) {
	err1 := customErrors.WrapSystemError(fmt.Errorf("a"))
	err2 := customErrors.WrapExternalServiceError(fmt.Errorf("b"))

	if errors.Is(err1, err2) {
		t.Error("errors with different base types should NOT match via errors.Is")
	}
}

func TestCustomError_Is_AgainstPlainSentinel(t *testing.T) {
	wrapped := customErrors.WrapSystemError(fmt.Errorf("inner"))

	if !errors.Is(wrapped, customErrors.ErrSystem) {
		t.Error("wrapped error should match its sentinel via errors.Is")
	}
	if errors.Is(wrapped, customErrors.ErrDataNotFound) {
		t.Error("wrapped system error should NOT match ErrDataNotFound sentinel")
	}
}

func TestIsSystemError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "plain error is a system error",
			err:      fmt.Errorf("unknown failure"),
			expected: true,
		},
		{
			name:     "wrapped system error",
			err:      customErrors.WrapSystemError(fmt.Errorf("sys")),
			expected: true,
		},
		{
			name:     "external service error is NOT a system error",
			err:      customErrors.WrapExternalServiceError(fmt.Errorf("ext")),
			expected: false,
		},
		{
			name:     "validation error is NOT a system error",
			err:      customErrors.WrapValidationError(fmt.Errorf("val")),
			expected: false,
		},
		{
			name:     "permission denied error is NOT a system error",
			err:      customErrors.WrapPermissionDeniedError(fmt.Errorf("perm")),
			expected: false,
		},
		{
			name:     "data not found error is NOT a system error",
			err:      customErrors.WrapDataNotFoundError(fmt.Errorf("nf")),
			expected: false,
		},
		{
			name:     "raw ErrDataNotFound sentinel is NOT a system error",
			err:      customErrors.ErrDataNotFound,
			expected: false,
		},
		{
			name:     "raw ErrExternalService sentinel is NOT a system error",
			err:      customErrors.ErrExternalService,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := customErrors.IsSystemError(tc.err)
			if got != tc.expected {
				t.Errorf("IsSystemError() = %v, want %v for error: %v", got, tc.expected, tc.err)
			}
		})
	}
}

func TestWrapChaining_DoubleWrapped(t *testing.T) {
	inner := fmt.Errorf("root cause")
	middle := customErrors.WrapSystemError(inner)
	outer := customErrors.WrapExternalServiceError(middle)

	// Outer wraps ErrExternalService
	if !errors.Is(outer, customErrors.ErrExternalService) {
		t.Error("outer error should match ErrExternalService")
	}
	// Inner error accessible via Unwrap chain
	if !errors.Is(outer, inner) {
		t.Error("root inner error should be reachable via Unwrap chain")
	}
}
