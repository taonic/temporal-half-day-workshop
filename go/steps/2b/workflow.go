package lab2

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

	// region: get-status
	if err := workflow.SetQueryHandler(ctx, QueryGetStatus, func() (TransferStatus, error) {
		return status, nil
	}); err != nil {
		return "", err
	}
	// endregion

	// region: approve-signal
	var approved, decided bool
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.GetSignalChannel(ctx, SignalApprove).Receive(ctx, &approved)
		decided = true
		logger.Info("Decision received", "approved", approved)
	})
	// endregion

	if err := workflow.ExecuteActivity(ctx, Withdraw, req.FromAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}

	// region: wait-for-approval
	status = StatusAwaitingApproval
	logger.Info("Waiting for approval")
	if err := workflow.Await(ctx, func() bool { return decided }); err != nil {
		return "", err
	}
	if !approved {
		if err := workflow.ExecuteActivity(ctx, Refund, req.FromAccount, req.Amount).Get(ctx, nil); err != nil {
			return "", err
		}
		status = StatusCancelled
		return status, nil
	}
	// endregion

	status = StatusApproved
	if err := workflow.ExecuteActivity(ctx, Deposit, req.ToAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}
	status = StatusCompleted
	logger.Info("Transfer completed", "transferId", req.TransferID)
	return status, nil
}
