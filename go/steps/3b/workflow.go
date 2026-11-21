package lab3

// region: imports
import (
	"errors"
	"fmt"
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
	toAccount := req.ToAccount
	var fixReceived bool
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
	// endregion

	// region: wait-for-handlers
	if err := workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) }); err != nil {
		return "", err
	}
	// endregion

	logger.Info("Transfer completed", "transferId", req.TransferID)
	return status, nil
}
