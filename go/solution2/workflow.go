package solution2

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// MoneyTransfer is registered under its function name, "MoneyTransfer".
func MoneyTransfer(ctx workflow.Context, req TransferRequest) (TransferStatus, error) {
	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    5,
		},
	})

	status := StatusPending

	// Query: read the current status without changing anything.
	if err := workflow.SetQueryHandler(ctx, QueryGetStatus, func() (TransferStatus, error) {
		return status, nil
	}); err != nil {
		return "", err
	}

	// Signal: a human approves (true) or rejects (false) the transfer.
	var approved, decided bool
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.GetSignalChannel(ctx, SignalApprove).Receive(ctx, &approved)
		decided = true
		logger.Info("Decision received", "approved", approved)
	})

	if err := workflow.ExecuteActivity(ctx, Withdraw, req.FromAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}

	status = StatusAwaitingApproval
	logger.Info("Waiting for approval", "timeout", ApprovalTimeout)

	// Durable timer: wait for the decision, or give up after ApprovalTimeout.
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
}
