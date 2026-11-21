---
slug: durable-timer
id: ld976wz1x8ki
type: challenge
title: '2c: A Durable Timer'
teaser: Give up after 30 seconds, even while no Worker is running.
tabs:
- id: lykcphvlogty
  title: Code Editor
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/go/lab2
  port: 8080
- id: ot5vivy6xtey
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: k4vrywqunsxe
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: 5okry8rtogu0
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: basic
timelimit: 900
enhanced_loading: null
---

# A Durable Timer

Nobody should wait forever. Refund the transfer if nobody approves it within
30 seconds, and watch the deadline hold while no Worker is running.

# Wait With a Deadline

`workflow.AwaitWithTimeout` is `Await` with a deadline. It returns `false`
when the 30 seconds ran out first. Behind it is a **durable timer**: the
server keeps it, not your process.

```go
status = StatusAwaitingApproval
logger.Info("Waiting for approval", "timeout", ApprovalTimeout)
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
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
apply wait-for-approval
```

# Run It, Then Stop the Worker

Restart the Worker in the [button label="Worker" background="#444CE7"](tab-1):
`Ctrl+C`, then:

```bash,run
go run ./lab2/worker
```

Start a transfer from the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab2/starter
```

Now stop the Worker: `Ctrl+C` in the
[button label="Worker" background="#444CE7"](tab-1). Nothing is running your
code. Ask the transfer where it is:

```bash,run
temporal workflow query --workflow-id money-transfer-workflow --name getStatus
```

It hangs, then fails, because a Worker answers Queries and none is running:

```bash,nocopy
Error: querying workflow failed: no poller seen for task queue recently, worker may be down
```

That took about 30 seconds, as long as the deadline. The timer has fired on
the server. Start the Worker again:

```bash,run
go run ./lab2/worker
```

It picks up the fired timer straight away and refunds. In the
[button label="Terminal" background="#444CE7"](tab-2):

```bash,run
temporal workflow result --workflow-id money-transfer-workflow
```

```bash,nocopy
Results:
  Status          COMPLETED
  Result          "TIMED_OUT"
  ResultEncoding  json/plain
```

# See It

In the [button label="Temporal UI" background="#444CE7"](tab-3):
`TimerStarted`, then `TimerFired` 30 seconds later, then the `Refund`.

# Predict, Then Try

The transfer has timed out. Approve it anyway:

```bash,run
temporal workflow signal --workflow-id money-transfer-workflow --name approve --input true
```

It fails: the Workflow has completed, and a finished Workflow takes no more
Signals.

Click **Check** once a transfer has timed out.

# Key Takeaways for Lab 2

- A Signal pushes data into a running Workflow. It does not wait for a reply.
- A Query reads state and writes nothing to the history. A Worker answers it.
- `AwaitWithTimeout` is a durable timer. It fires even when every Worker is down.
- A stopped Worker pauses progress but loses nothing. The history is the state.
