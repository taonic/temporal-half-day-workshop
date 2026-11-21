# Lab 3 Solution Snippets

Paste each snippet over its `TODO` comment, then delete the comment. The full
files are in [`../solution3/`](../solution3/).

## TODO 1, `activities.go`: a non-retryable error

```go
	if account == InvalidAccount {
		return temporal.NewNonRetryableApplicationError(
			"deposit failed: account "+account+" does not exist",
			ErrTypeInvalidAccount,
			nil,
		)
	}
```

## TODO 2, `workflow.go`: the `updateToAccount` Update

```go
	if err := workflow.SetUpdateHandlerWithOptions(ctx, UpdateToAccount,
		func(ctx workflow.Context, account string) (TransferStatus, error) {
			toAccount = account
			fixReceived = true
			if err := workflow.Await(ctx, func() bool { return status != StatusRetrying }); err != nil {
				return "", err
			}
			return status, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, account string) error {
				if status != StatusRetrying {
					return fmt.Errorf("transfer is %s, not waiting for a corrected account", status)
				}
				if account == "" || account == InvalidAccount {
					return fmt.Errorf("%q is not a valid account", account)
				}
				return nil
			},
		},
	); err != nil {
		return "", err
	}
```

## TODO 3, `workflow.go`: the deposit loop

Replace the single `Deposit` call and the `status = StatusCompleted` after it:

```go
	for {
		err := workflow.ExecuteActivity(ctx, Deposit, toAccount, req.Amount).Get(ctx, nil)
		if err == nil {
			break
		}

		var appErr *temporal.ApplicationError
		if !errors.As(err, &appErr) || appErr.Type() != ErrTypeInvalidAccount {
			status = StatusFailed
			return status, err
		}

		logger.Warn("Deposit rejected, waiting for updateToAccount", "account", toAccount)
		status = StatusRetrying
		fixReceived = false
		if err := workflow.Await(ctx, func() bool { return fixReceived }); err != nil {
			return "", err
		}
	}
	status = StatusCompleted
```

## TODO 4, `workflow.go`: wait for handlers

```go
	if err := workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) }); err != nil {
		return "", err
	}
```

## TODO 5, `workflow_test.go`: the timeout path

```go
	env.OnActivity(Withdraw, mock.Anything, "account-123", 100.0).Return(nil)
	env.OnActivity(Refund, mock.Anything, "account-123", 100.0).Return(nil)

	env.ExecuteWorkflow(MoneyTransfer, request("account-456"))

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var status TransferStatus
	require.NoError(t, env.GetWorkflowResult(&status))
	require.Equal(t, StatusTimedOut, status)
	env.AssertActivityNumberOfCalls(t, "Refund", 1)
```

## TODO 6, `workflow_test.go`: the invalid account fixed by an Update

```go
	env.OnActivity(Withdraw, mock.Anything, "account-123", 100.0).Return(nil)
	env.OnActivity(Deposit, mock.Anything, InvalidAccount, 100.0).
		Return(temporal.NewNonRetryableApplicationError("no such account", ErrTypeInvalidAccount, nil))
	env.OnActivity(Deposit, mock.Anything, "account-456", 100.0).Return(nil)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalApprove, true)
	}, time.Second)

	var updateResult TransferStatus
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateToAccount, "fix-1", &testsuite.TestUpdateCallback{
			OnReject: func(err error) { t.Errorf("update rejected: %v", err) },
			OnComplete: func(result interface{}, err error) {
				require.NoError(t, err)
				updateResult = result.(TransferStatus)
			},
		}, "account-456")
	}, 2*time.Second)

	env.ExecuteWorkflow(MoneyTransfer, request(InvalidAccount))

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var status TransferStatus
	require.NoError(t, env.GetWorkflowResult(&status))
	require.Equal(t, StatusCompleted, status)
	require.Equal(t, StatusCompleted, updateResult)
```
