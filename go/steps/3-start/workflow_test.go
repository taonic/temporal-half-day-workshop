package lab3

// region: imports
import (
	"testing"

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
	t.Skip("Page 3c writes this test.")
	// endregion
}

func TestInvalidAccountFixedByUpdate(t *testing.T) {
	// region: test-update
	t.Skip("Page 3c writes this test.")
	// endregion
}
