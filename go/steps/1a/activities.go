package lab1

// region: imports
import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
)

// endregion

// ComposeGreeting stands in for a call to a greeting service.
func ComposeGreeting(ctx context.Context, name string) (string, error) {
	// region: flaky-service
	// The greeting service never fails yet. Page 1b changes that.
	// endregion

	// region: compose-greeting
	activity.GetLogger(ctx).Info("Composing greeting", "name", name, "attempt", activity.GetInfo(ctx).Attempt)
	return fmt.Sprintf("Hello, %s!", name), nil
	// endregion
}
