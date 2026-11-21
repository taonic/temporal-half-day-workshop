# Lab 1 Solution Snippets

Paste each snippet over its `TODO` comment, then delete the comment. The full
files are in [`../solution1/`](../solution1/).

## `activities.go`: the success path

```go
	return fmt.Sprintf("Hello, %s!", name), nil
```

## `workflow.go`: Activity options with a retry policy

```go
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
```

## `workflow.go`: execute the Activity

```go
	var greeting string
	err := workflow.ExecuteActivity(ctx, ComposeGreeting, name).Get(ctx, &greeting)
	if err != nil {
		return "", err
	}
	return greeting, nil
```

## `worker/main.go`: register

```go
	w.RegisterWorkflow(lab1.Greeting)
	w.RegisterActivity(lab1.ComposeGreeting)
```

## `starter/main.go`: start the Workflow

```go
	we, err := c.ExecuteWorkflow(context.Background(), options, lab1.Greeting, "Temporal")
```
