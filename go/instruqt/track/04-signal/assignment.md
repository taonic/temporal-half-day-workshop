---
slug: signal
type: challenge
title: '2a: Wait for a Signal'
teaser: A money transfer that waits, as long as it takes, for a person to approve it.
notes:
- type: text
  contents: |-
    # How long can a function wait for a person?

    A person approves the transfer in five seconds, or after lunch, or
    never. A normal program cannot hold a thread for that, and cannot
    survive a restart while it does.

    A Workflow can. Lab 2 builds a money transfer that waits for a person.
- type: text
  contents: |-
    # Three ways to talk to a Workflow

    **Signal** (this page): send it data. It does not reply.

    **Query** (page 2b): ask it for its state. It changes nothing.

    **Timer** (page 2c): wait a set time, kept by the server. It fires even
    when every Worker is down.
tabs:
- title: Code Editor
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/go/lab2
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
difficulty: basic
timelimit: 900
enhanced_loading: null
---

# Wait for a Signal

`lab2` holds a money transfer that withdraws, then deposits straight away.
Make it wait for a person to approve.

The Activities, Worker and starter are done. Every change is in
`workflow.go`. Open it in the
[button label="Code Editor" background="#444CE7"](tab-0) to see where the
transfer is now.

# 1. Receive the Signal

A Signal arrives on a channel. This goroutine waits for one `approve`
Signal with a `bool` and records the decision. Use `workflow.Go`, not `go`:
Temporal schedules Workflow goroutines so it can replay them.

```go
var approved, decided bool
workflow.Go(ctx, func(ctx workflow.Context) {
    workflow.GetSignalChannel(ctx, SignalApprove).Receive(ctx, &approved)
    decided = true
    logger.Info("Decision received", "approved", approved)
})
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
apply approve-signal
```

# 2. Wait for the Decision

`workflow.Await` blocks until the condition is true. It is checked again
each time something changes, such as a Signal arriving. Rejected? Refund
the money and stop.

```go
status = StatusAwaitingApproval
logger.Info("Waiting for approval")
if err := workflow.Await(ctx, func() bool { return decided }); err != nil {
    return "", err
}
if !approved {
    if err := workflow.ExecuteActivity(ctx, Refund, req.FromAccount, req.Amount).Get(ctx, nil); err != nil {
        return "", err
    }
    status = StatusCancelled
    return status, nil
}
```

```bash,run
apply wait-for-approval
```

# Run It

Start the Worker in the [button label="Worker" background="#444CE7"](tab-1):

```bash,run
go run ./lab2/worker
```

Start a transfer from the [button label="Terminal" background="#444CE7"](tab-2).
The starter returns at once; the transfer keeps running on the server:

```bash,run
go run ./lab2/starter
```

The transfer has withdrawn the money and is waiting. Take your time. Then
approve it:

```bash,run
temporal workflow signal --workflow-id money-transfer-workflow --name approve --input true
```

Ask for the result:

```bash,run
temporal workflow result --workflow-id money-transfer-workflow
```

```bash,nocopy
Results:
  Status          COMPLETED
  Result          "COMPLETED"
  ResultEncoding  json/plain
```

# See It

In the [button label="Temporal UI" background="#444CE7"](tab-3), open
`money-transfer-workflow`. `WorkflowExecutionSignaled` sits between the
withdrawal and the deposit, with your `true` as its input.

# Predict, Then Try

Run the starter again and send `--input false` instead. What is the result,
and which Activities ran?

`CANCELLED`, after `Withdraw` and then `Refund`. No deposit.

Click **Check** once a transfer has completed through an approve Signal.
