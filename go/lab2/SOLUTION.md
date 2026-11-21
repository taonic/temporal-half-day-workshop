# Lab 2 Solution Snippets

Every TODO is in `workflow.go`. Paste each snippet over its `TODO` comment,
then delete the comment. The full file is
[`../solution2/workflow.go`](../solution2/workflow.go).

## TODO 1: the `getStatus` Query

```go
	if err := workflow.SetQueryHandler(ctx, QueryGetStatus, func() (TransferStatus, error) {
		return status, nil
	}); err != nil {
		return "", err
	}
```

## TODO 2: the `approve` Signal

```go
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.GetSignalChannel(ctx, SignalApprove).Receive(ctx, &approved)
		decided = true
		logger.Info("Decision received", "approved", approved)
	})
```

## TODO 3: the approval deadline

```go
	gotDecision, err := workflow.AwaitWithTimeout(ctx, ApprovalTimeout, func() bool { return decided })
	if err != nil {
		return "", err
	}

	if !gotDecision || !approved {
		if err := workflow.ExecuteActivity(ctx, Refund, req.FromAccount, req.Amount).Get(ctx, nil); err != nil {
			return "", err
		}
		if !gotDecision {
			status = StatusTimedOut
		} else {
			status = StatusCancelled
		}
		return status, nil
	}

	status = StatusApproved
	if err := workflow.ExecuteActivity(ctx, Deposit, req.ToAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}
	status = StatusCompleted
	return status, nil
```
