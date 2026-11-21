package lab3

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// MoneyTransfer is registered under its function name, "MoneyTransfer".
// This starts as the Lab 2 solution. This lab changes how the deposit step fails.
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

	// TODO 2 (Update): Register an Update handler named UpdateToAccount with
	//        workflow.SetUpdateHandlerWithOptions(ctx, UpdateToAccount, handler, options).
	//
	//        handler func(ctx workflow.Context, account string) (TransferStatus, error):
	//          set toAccount = account and fixReceived = true, then
	//          workflow.Await until status != StatusRetrying, and return status.
	//
	//        options workflow.UpdateHandlerOptions{Validator: ...}, where the
	//        validator func(ctx workflow.Context, account string) error rejects:
	//          - any call while status != StatusRetrying
	//          - an empty account, or account == InvalidAccount
	//        Use fmt.Errorf for the rejection messages.

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

	// TODO 3 (Deposit loop): Replace the single Deposit call below with a
	//        for loop that deposits to toAccount:
	//        - success: break out of the loop
	//        - an InvalidAccount error: set status = StatusRetrying and
	//          fixReceived = false, workflow.Await until fixReceived, then loop
	//        - any other error: set status = StatusFailed and return it with err
	//        Detect the error type with:
	//          var appErr *temporal.ApplicationError
	//          errors.As(err, &appErr) && appErr.Type() == ErrTypeInvalidAccount
	if err := workflow.ExecuteActivity(ctx, Deposit, toAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}
	status = StatusCompleted

	// TODO 4 (Handlers): Before returning, let the Update handler send its
	//        result back:
	//        workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) })

	return status, nil
}
