---
slug: testing
type: challenge
title: '3c: Test It, Skipping Time'
teaser: Test the 30-second timeout and the Update fix in milliseconds, with no server.
tabs:
- title: Code Editor
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/go/lab3
  port: 8080
- title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: intermediate
timelimit: 900
enhanced_loading: null
---

# Test It, Skipping Time

`workflow_test.go` has its setup done: `newEnv` builds an in-memory test
environment that **skips time**. A 30-second timer fires as soon as nothing
else is left to run. Two tests are waiting to be written.

# 1. The Timeout Path

Mock the Activities this path calls, send no Signal, and check the result.
`OnActivity` replaces a real Activity for the arguments given.

```go
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
```

# 2. The Update Fix

The mocked `Deposit` fails for `invalid-account` and succeeds for
`account-456`. Delayed callbacks play the person: approve at 1 second, send
the Update at 2 seconds, both in skipped time.

```go
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
```

In the [button label="Terminal" background="#444CE7"](tab-2), apply both:

```bash,run
apply test-timeout test-update
```

No Worker restart: tests build their own environment.

# Run Them

```bash,run
go test -v ./lab3
```

```bash,nocopy
--- PASS: TestApprovalTimesOut (0.02s)
--- PASS: TestInvalidAccountFixedByUpdate (0.00s)
PASS
```

`TestApprovalTimesOut` waited 30 seconds of Workflow time in a few
milliseconds. Neither test touched the Temporal server.

# Predict, Then Try

In `TestApprovalTimesOut`, change the `Refund` mock's account from
`"account-123"` to `"account-999"`. Run the tests. What fails, and how?

The Workflow refunds `account-123`, which no mock expects, so the test
fails and says which Activity call it could not match. Mocks check the
arguments, not only the names.

Click **Check** once both tests pass.

# Key Takeaways for Lab 3

- A non-retryable error skips the retry policy and goes straight to the Workflow.
- A Workflow can park on bad data and resume when a person fixes it. Completed steps stay completed.
- An Update has a validator, so bad requests never reach history, and a handler that returns a result.
- Wait for `AllHandlersFinished` before returning, so an in-flight Update gets its reply.
- The test environment skips time and mocks Activities, so timer and failure paths test in milliseconds.
