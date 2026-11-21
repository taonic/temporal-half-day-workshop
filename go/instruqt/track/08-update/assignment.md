---
slug: update
id: g1llojdaa0h4
type: challenge
title: '3b: Fix It With an Update'
teaser: The transfer parks on the bad account until a person sends the right one.
notes:
- type: text
  contents: |-
    # Signal or Update?

    A **Signal** is fire-and-forget: the sender learns nothing.

    An **Update** is a request with a reply. A validator can reject it
    before anything is written to history, and the handler can wait until
    the work is done and return the result.

    Fixing a bank account deserves a reply, so this page uses an Update.
tabs:
- id: enmrbvrt23mh
  title: Code Editor
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/go/lab3
  port: 8080
- id: 63tf5xz6eytn
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: vhiokekoskzb
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: gqlmevt93ngp
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: intermediate
timelimit: 900
enhanced_loading: null
---

# Fix It With an Update

Instead of failing, the transfer parks in `RETRYING` and waits for an
`updateToAccount` Update with a corrected account. Every change is in
`workflow.go`.

# 1. The Update Handler

The **validator** decides whether to accept the request, and must not
change anything. It rejects a bad account, or any request while the transfer
is not waiting for one. The **handler** stores the new account, wakes the
deposit loop, then waits until the deposit is done and returns the outcome.

```go
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
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
apply update-to-account
```

# 2. Park Instead of Failing

The deposit becomes a loop. On an `InvalidAccount` error it sets the status
to `RETRYING` and waits for the handler to deliver a fix, then tries again.
Any other error still fails the transfer.

```go
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
```

```bash,run
apply deposit
```

# 3. Let the Reply Get Out

The Workflow must not finish while the handler is still sending its reply.
Wait for every handler first:

```go
if err := workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) }); err != nil {
    return "", err
}
```

```bash,run
apply wait-for-handlers
```

# Run It

Restart the Worker in the [button label="Worker" background="#444CE7"](tab-1):
`Ctrl+C`, then:

```bash,run
go run ./lab3/worker
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab3/starter
```

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
temporal workflow update execute --workflow-id money-transfer-workflow --name updateToAccount --input '"invalid-account"'
```

```bash,nocopy
Error: unable to update workflow: "invalid-account" is not a valid account
```

The validator rejected it. Check the
[button label="Temporal UI" background="#444CE7"](tab-3): the rejected Update
left no event at all.

# Send the Real Fix

```bash,run
temporal workflow update execute --workflow-id money-transfer-workflow --name updateToAccount --input '"account-456"'
```

```bash,nocopy
  Name      updateToAccount
  UpdateID  ...
  Result    COMPLETED
```

The reply came back after the retried deposit finished. A Signal cannot
tell you that.

# See It

In the history: `ActivityTaskFailed` with `nonRetryable: true`, then
`WorkflowExecutionUpdateAccepted`, the second `Deposit` to `account-456`,
and `WorkflowExecutionUpdateCompleted`. The withdrawal ran once and stayed
done.

# Predict, Then Try

The transfer has completed. Send the real fix again. What happens?

It fails: a finished Workflow accepts no Updates. The validator never even
runs.

Click **Check** once a transfer has completed through `updateToAccount`.
