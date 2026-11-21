package lab3

// region: imports
import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

// endregion

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
	// region: test-timeout
	env := newEnv(t)
	env.OnActivity(Withdraw, mock.Anything, "account-123", 100.0).Return(nil)
	env.OnActivity(Refund, mock.Anything, "account-123", 100.0).Return(nil)

	// No Signal is sent, so the 30 second deadline passes, in skipped time.
	env.ExecuteWorkflow(MoneyTransfer, request("account-456"))

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var status TransferStatus
	require.NoError(t, env.GetWorkflowResult(&status))
	require.Equal(t, StatusTimedOut, status)
	env.AssertActivityNumberOfCalls(t, "Refund", 1)
	// endregion
}

func TestInvalidAccountFixedByUpdate(t *testing.T) {
	// region: test-update
	env := newEnv(t)
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
	// endregion
}
