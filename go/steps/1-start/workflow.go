package lab1

// region: imports
import "go.temporal.io/sdk/workflow"

// endregion

const TaskQueue = "greeting-task-queue"

// Greeting is registered under its function name, "Greeting".
func Greeting(ctx workflow.Context, name string) (string, error) {
	// region: activity-options
	// Page 1a sets the Activity options.
	// endregion

	// region: execute-activity
	// Page 1a calls the Activity.
	return "", nil
	// endregion
}
