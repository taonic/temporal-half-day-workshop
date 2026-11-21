---
slug: failures-and-testing
type: challenge
title: 'Lab 3: Failure Handling and Testing'
teaser: A bad account number parks the transfer until a human fixes it with an Update.
  Then prove it with tests that skip time.
notes:
- type: text
  contents: |-
    # Some failures should not be retried

    A network blip is worth retrying. A wrong account number is not:
    retrying it five times gets you the same error five times.

    Mark that error **non-retryable**. The Workflow then gets it at once,
    and can park and wait for a human to fix the data. The withdrawal that
    already went through stays done.
- type: text
  contents: |-
    # Signal or Update?

    A **Signal** is fire-and-forget: the sender learns nothing.

    An **Update** is a request with a reply. A validator can reject it
    before anything is written to history, and the handler can wait until
    the work is done and return the result.

    Fixing a bank account should get a reply, so this lab uses an Update.
tabs:
- title: Code Editor
  type: code
  hostname: workshop
  path: /root/workshop/go/lab3
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
- title: Solution
  type: code
  hostname: workshop
  path: /root/workshop/go/solution3
- title: CLI
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
difficulty: intermediate
timelimit: 1800
enhanced_loading: null
---

# Failure Handling and Testing

The Lab 2 transfer, with one change: `Deposit` rejects `invalid-account` with
an error Temporal will not retry. The Workflow parks in `RETRYING` until an
`updateToAccount` Update sends a corrected account.

# The Code

Open the [button label="Code Editor" background="#444CE7"](tab-0). The TODOs
are numbered in order:

| TODO | File | What it does |
|------|------|--------------|
| 1 | `activities.go` | `Deposit` returns a **non-retryable** `InvalidAccount` error for `invalid-account` |
| 2 | `workflow.go` | The `updateToAccount` handler and its validator |
| 3 | `workflow.go` | A deposit loop that parks in `RETRYING` on `InvalidAccount` |
| 4 | `workflow.go` | Wait for the Update handler to finish before returning |
| 5, 6 | `workflow_test.go` | Two tests: the timeout path, and the invalid account fixed by an Update |

The non-retryable error is one call:

```go
return temporal.NewNonRetryableApplicationError(
    "deposit failed: account "+account+" does not exist",
    ErrTypeInvalidAccount,
    nil,
)
```

The Update has two halves. The **validator** decides whether to accept the
request, and it must not change anything. The **handler** does the work and
returns the result:

```go
workflow.SetUpdateHandlerWithOptions(ctx, UpdateToAccount,
    func(ctx workflow.Context, account string) (TransferStatus, error) { /* handler */ },
    workflow.UpdateHandlerOptions{
        Validator: func(ctx workflow.Context, account string) error { /* validator */ },
    },
)
```

> [!IMPORTANT]
> Delete each `// TODO` comment as you finish. `lab3/SOLUTION.md` has every
> snippet.

# Run It

Start the Worker in the [button label="Worker" background="#444CE7"](tab-1):

```bash,run
go run ./lab3/worker
```

Start the transfer in the [button label="Terminal" background="#444CE7"](tab-2).
The starter deposits to `invalid-account` and approves after two seconds:

```bash,run
go run ./lab3/starter
```

In the [button label="CLI" background="#444CE7"](tab-5), check where it parked:

```bash,run
temporal workflow query --workflow-id money-transfer-workflow --name getStatus
```

```bash,nocopy
Query result:
  QueryResult  "RETRYING"
```

# Try a Bad Fix

Send the same bad account as the fix:

```bash,run
temporal workflow update execute --workflow-id money-transfer-workflow \
  --name updateToAccount --input '"invalid-account"'
```

```bash,nocopy
Error: unable to update workflow: "invalid-account" is not a valid account
```

The validator rejected it. Look at the history in the
[button label="Temporal UI" background="#444CE7"](tab-3): the rejected Update
left no event at all.

# Send the Real Fix

```bash,run
temporal workflow update execute --workflow-id money-transfer-workflow \
  --name updateToAccount --input '"account-456"'
```

```bash,nocopy
  Result    COMPLETED
```

The Update waited for the retried deposit and returned the outcome. A Signal
cannot do that. The starter prints `Final status: COMPLETED`.

In the history, find the `ActivityTaskFailed` with `nonRetryable: true` and no
second attempt, then `WorkflowExecutionUpdateAccepted` and
`WorkflowExecutionUpdateCompleted`.

# Test It

Fill in TODOs 5 and 6 in `workflow_test.go`. The environment is already set
up, and it **skips time**: the 30-second approval deadline passes instantly.

```bash,run
go test -v ./lab3
```

```bash,nocopy
--- PASS: TestApprovalTimesOut (0.04s)
--- PASS: TestInvalidAccountFixedByUpdate (0.00s)
PASS
```

`TestApprovalTimesOut` waited 30 seconds of Workflow time in well under a
second. `TestInvalidAccountFixedByUpdate` mocks `Deposit`, so neither test
touches a server.

Click **Check** when the tests pass and a transfer has completed through
`updateToAccount`.

# Key Takeaways

- A non-retryable error skips the retry policy and goes straight to the Workflow.
- A Workflow can park on bad data and resume when a human fixes it. Completed steps stay completed.
- An Update has a validator, so bad requests never reach history, and a handler that returns a result.
- Wait for `AllHandlersFinished` before returning, so an in-flight Update gets its reply.
- The test environment skips time and mocks Activities, so timer and failure paths test in milliseconds.
