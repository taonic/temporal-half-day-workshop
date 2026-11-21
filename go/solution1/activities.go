package solution1

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// ComposeGreeting stands in for a flaky greeting service. It fails on its
// first two attempts and succeeds on the third, so the retry policy in the
// Workflow is what gets the greeting through.
func ComposeGreeting(ctx context.Context, name string) (string, error) {
	attempt := activity.GetInfo(ctx).Attempt
	activity.GetLogger(ctx).Info("Composing greeting", "name", name, "attempt", attempt)

	if attempt < 3 {
		return "", temporal.NewApplicationError(
			fmt.Sprintf("greeting service unavailable (attempt %d)", attempt),
			"ServiceUnavailable",
		)
	}

	return fmt.Sprintf("Hello, %s!", name), nil
}
