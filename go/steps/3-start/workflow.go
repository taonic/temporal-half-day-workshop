package lab3

// region: imports
import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// endregion

// MoneyTransfer is registered under its function name, "MoneyTransfer".
// It starts as the finished Lab 2 transfer. This lab changes what happens
// when the deposit fails.
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

	if err := workflow.SetQueryHandler(ctx, QueryGetStatus, func() (TransferStatus, error) {
		return status, nil
	}); err != nil {
		return "", err
	}

	var approved, decided bool
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.GetSignalChannel(ctx, SignalApprove).Receive(ctx, &approved)
		decided = true
		logger.Info("Decision received", "approved", approved)
	})

	// region: update-to-account
	// Page 3b lets a person fix the account here, with an Update.
	// endregion

	if err := workflow.ExecuteActivity(ctx, Withdraw, req.FromAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}

	status = StatusAwaitingApproval
	logger.Info("Waiting for approval", "timeout", ApprovalTimeout)
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

	// region: deposit
	if err := workflow.ExecuteActivity(ctx, Deposit, req.ToAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}
	status = StatusCompleted
	// endregion

	// region: wait-for-handlers
	// Page 3b waits here for the Update handler to finish.
	// endregion

	logger.Info("Transfer completed", "transferId", req.TransferID)
	return status, nil
}
