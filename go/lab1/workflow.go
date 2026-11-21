package lab1

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const TaskQueue = "greeting-task-queue"

// Greeting is registered under its function name, "Greeting".
func Greeting(ctx workflow.Context, name string) (string, error) {
	// TODO: Build workflow.ActivityOptions with:
	//       - StartToCloseTimeout: 5 seconds
	//       - RetryPolicy: &temporal.RetryPolicy{} with InitialInterval 1 second,
	//         BackoffCoefficient 2.0 and MaximumAttempts 5
	//       Then apply them: ctx = workflow.WithActivityOptions(ctx, ao)

	// TODO: Execute ComposeGreeting with name, wait for the result with
	//       .Get(ctx, &greeting), and return the greeting.
}
