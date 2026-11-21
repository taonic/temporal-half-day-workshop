package lab1

// region: imports
import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// endregion

// ComposeGreeting stands in for a call to a greeting service.
func ComposeGreeting(ctx context.Context, name string) (string, error) {
	// region: flaky-service
	attempt := activity.GetInfo(ctx).Attempt
	if attempt < 3 {
		return "", temporal.NewApplicationError(
			fmt.Sprintf("greeting service unavailable (attempt %d)", attempt),
			"ServiceUnavailable",
		)
	}
	// endregion

	// region: compose-greeting
	activity.GetLogger(ctx).Info("Composing greeting", "name", name, "attempt", activity.GetInfo(ctx).Attempt)
	return fmt.Sprintf("Hello, %s!", name), nil
	// endregion
}
