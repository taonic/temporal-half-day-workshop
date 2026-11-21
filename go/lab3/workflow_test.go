package lab3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

// newEnv returns a test environment with the Workflow registered. Activities
// are not registered: every test mocks the ones it expects with OnActivity.
//
// The environment skips time automatically. A 30 second timer fires as soon
// as nothing else is left to run.
func newEnv(t *testing.T) *testsuite.TestWorkflowEnvironment {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(MoneyTransfer)
	t.Cleanup(func() { env.AssertExpectations(t) })
	return env
}

func request(toAccount string) TransferRequest {
	return TransferRequest{
		FromAccount: "account-123",
		ToAccount:   toAccount,
		Amount:      100,
		TransferID:  "transfer-test",
	}
}

func TestApprovalTimesOut(t *testing.T) {
	env := newEnv(t)

	// TODO 5 (Test the timer):
	//   1. Mock Withdraw and Refund for "account-123", 100.0 to return nil:
	//      env.OnActivity(Withdraw, mock.Anything, "account-123", 100.0).Return(nil)
	//   2. Run env.ExecuteWorkflow(MoneyTransfer, request("account-456"))
	//      and send no Signal. The 30 second deadline passes in skipped time.
	//   3. Require the Workflow completed without error, its result is
	//      StatusTimedOut, and env.AssertActivityNumberOfCalls(t, "Refund", 1).
}

func TestInvalidAccountFixedByUpdate(t *testing.T) {
	env := newEnv(t)

	// TODO 6 (Test the Update):
	//   1. Mock Withdraw to return nil. Mock Deposit for InvalidAccount to return
	//      temporal.NewNonRetryableApplicationError("no such account", ErrTypeInvalidAccount, nil),
	//      and Deposit for "account-456" to return nil.
	//   2. env.RegisterDelayedCallback: after 1 second, send the approve Signal
	//      with env.SignalWorkflow(SignalApprove, true).
	//   3. env.RegisterDelayedCallback: after 2 seconds, call
	//      env.UpdateWorkflow(UpdateToAccount, "fix-1", &testsuite.TestUpdateCallback{...}, "account-456").
	//      Fail the test in OnReject. In OnComplete, save result.(TransferStatus).
	//   4. Run env.ExecuteWorkflow(MoneyTransfer, request(InvalidAccount)).
	//   5. Require the Workflow result and the saved Update result are both
	//      StatusCompleted.
}
