package solution1

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const TaskQueue = "greeting-task-queue"

// Greeting is registered under its function name, "Greeting".
func Greeting(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var greeting string
	err := workflow.ExecuteActivity(ctx, ComposeGreeting, name).Get(ctx, &greeting)
	if err != nil {
		return "", err
	}
	return greeting, nil
}
