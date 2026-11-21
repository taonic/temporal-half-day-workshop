---
slug: non-retryable-error
id: mj6giuuc3wtj
type: challenge
title: '3a: An Error Not Worth Retrying'
teaser: A bad account number fails the deposit for good, at once.
notes:
- type: text
  contents: |-
    # Some failures should not be retried

    A network blip is worth retrying. A wrong account number is not:
    retrying it five times gets you the same error five times.

    Mark that error **non-retryable**, and Temporal hands it to the
    Workflow at once.
tabs:
- id: 95fznujtxu14
  title: Code Editor
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/go/lab3
  port: 8080
- id: 02aeggkrz0vt
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: vnbavvgftqwy
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: 2fbspckisex0
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: intermediate
timelimit: 900
enhanced_loading: null
---

# An Error Not Worth Retrying

`lab3` holds the finished Lab 2 transfer. Its starter deposits to
`invalid-account` and approves after two seconds. Make `Deposit` reject that
account in a way Temporal will not retry.

# The Non-Retryable Error

`NewNonRetryableApplicationError` skips the retry policy. The error type,
`InvalidAccount`, lets the Workflow tell this failure apart from others. In
`activities.go`:

```go
if account == InvalidAccount {
    return temporal.NewNonRetryableApplicationError(
        "deposit failed: account "+account+" does not exist",
        ErrTypeInvalidAccount,
        nil,
    )
}
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
apply invalid-account
```

# Run It

Start the Worker in the [button label="Worker" background="#444CE7"](tab-1):

```bash,run
go run ./lab3/worker
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab3/starter
```

```bash,nocopy
Started money-transfer-workflow with a deposit to invalid-account
Approved
```

```bash,run
temporal workflow result --workflow-id money-transfer-workflow
```

```bash,nocopy
Results:
  Status   FAILED
  Failure
    Message: activity error
    Cause:
        Message: deposit failed: account invalid-account does not exist
Error: workflow failed
```

# See It

In the [button label="Temporal UI" background="#444CE7"](tab-3), find the
`ActivityTaskFailed` for `Deposit`: `nonRetryable: true`, and no second
attempt.

Look at what ran before it: `Withdraw` completed. The money left
`account-123` and went nowhere, and the Workflow is over. Page 3b fixes
that: the transfer will wait for a person to correct the account.

# Predict, Then Try

In `activities.go`, change `NewNonRetryableApplicationError(` to
`NewApplicationError(` and delete the `nil,` line, making the error
retryable. Restart the Worker, run the starter, and ask for the result. How
long does it take to fail?

About 15 seconds: five attempts, with waits of 1, 2, 4 and 8 seconds. Every
attempt fails the same way.

Click **Check** once a transfer has failed on the non-retryable error.
