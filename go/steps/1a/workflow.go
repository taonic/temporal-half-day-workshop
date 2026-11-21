package lab1

// region: imports
import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// endregion

const TaskQueue = "greeting-task-queue"

// Greeting is registered under its function name, "Greeting".
func Greeting(ctx workflow.Context, name string) (string, error) {
	// region: activity-options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	// endregion

	// region: execute-activity
	var greeting string
	err := workflow.ExecuteActivity(ctx, ComposeGreeting, name).Get(ctx, &greeting)
	if err != nil {
		return "", err
	}
	return greeting, nil
	// endregion
}
