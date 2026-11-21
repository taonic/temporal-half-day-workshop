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

	// TODO 1 (Query): Register a Query handler named QueryGetStatus that
	//        returns the current status:
	//        workflow.SetQueryHandler(ctx, QueryGetStatus, func() (TransferStatus, error) { ... })
	//        Return the error if registration fails.

	var approved, decided bool

	// TODO 2 (Signal): In a workflow.Go goroutine, receive one bool from
	//        workflow.GetSignalChannel(ctx, SignalApprove) into approved,
	//        then set decided = true.

	if err := workflow.ExecuteActivity(ctx, Withdraw, req.FromAccount, req.Amount).Get(ctx, nil); err != nil {
		return "", err
	}

	status = StatusAwaitingApproval
	logger.Info("Waiting for approval", "timeout", ApprovalTimeout)

	// TODO 3 (Timer): Wait for decided, for at most ApprovalTimeout:
	//        gotDecision, err := workflow.AwaitWithTimeout(ctx, ApprovalTimeout, func() bool { return decided })
	//        Then handle the three outcomes:
	//        - no decision in time: run Refund, status = StatusTimedOut,  return status
	//        - rejected (false):    run Refund, status = StatusCancelled, return status
	//        - approved (true):     status = StatusApproved, run Deposit to req.ToAccount,
	//                               status = StatusCompleted, return status
	//        Refund goes back to req.FromAccount. Call each Activity the
	//        way Withdraw is called above.
}
