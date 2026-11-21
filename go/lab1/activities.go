package lab1

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

	// The flaky part is done for you. Leave it in.
	if attempt < 3 {
		return "", temporal.NewApplicationError(
			fmt.Sprintf("greeting service unavailable (attempt %d)", attempt),
			"ServiceUnavailable",
		)
	}

	// TODO: Return the greeting "Hello, <name>!" and a nil error.
	//       fmt.Sprintf("Hello, %s!", name) builds the string.
}
