package solution3

import (
	"errors"
	"fmt"
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
	toAccount := req.ToAccount
	var fixReceived bool

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

	// Update: correct the deposit account and get the outcome back.
	if err := workflow.SetUpdateHandlerWithOptions(ctx, UpdateToAccount,
		func(ctx workflow.Context, account string) (TransferStatus, error) {
			toAccount = account
			fixReceived = true
			// Wait for the deposit loop to retry with the new account.
			if err := workflow.Await(ctx, func() bool { return status != StatusRetrying }); err != nil {
				return "", err
			}
			return status, nil
		},
		workflow.UpdateHandlerOptions{
			// The validator runs before the Update is written to history.
			// A rejected Update leaves no trace in the Workflow.
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
	for {
		err := workflow.ExecuteActivity(ctx, Deposit, toAccount, req.Amount).Get(ctx, nil)
		if err == nil {
			break
		}

		var appErr *temporal.ApplicationError
		if !errors.As(err, &appErr) || appErr.Type() != ErrTypeInvalidAccount {
			// Transient errors were already retried by the retry policy.
			status = StatusFailed
			return status, err
		}

		// Retrying cannot fix a bad account. Park until a human sends one.
		logger.Warn("Deposit rejected, waiting for updateToAccount", "account", toAccount)
		status = StatusRetrying
		fixReceived = false
		if err := workflow.Await(ctx, func() bool { return fixReceived }); err != nil {
			return "", err
		}
	}
	status = StatusCompleted

	// Let the Update handler return its result before the Workflow closes.
	if err := workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) }); err != nil {
		return "", err
	}
	return status, nil
}
