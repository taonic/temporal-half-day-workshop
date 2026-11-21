package lab1

// region: imports
import "context"

// endregion

// ComposeGreeting stands in for a call to a greeting service.
func ComposeGreeting(ctx context.Context, name string) (string, error) {
	// region: flaky-service
	// The greeting service never fails yet. Page 1b changes that.
	// endregion

	// region: compose-greeting
	// Page 1a writes the greeting.
	return "", nil
	// endregion
}
