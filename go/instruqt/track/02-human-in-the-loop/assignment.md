---
slug: human-in-the-loop
type: challenge
title: 'Lab 2: Human in the Loop'
teaser: A transfer waits for a human to approve it, gives up after 30 seconds, and
  tells you where it is while it waits.
notes:
- type: text
  contents: |-
    # How long can a function wait for a person?

    A person approves the transfer in five seconds, or after lunch, or
    never. A normal program cannot hold a thread for that, and cannot
    survive a restart while it does.

    A Workflow can. It waits for a Signal, gives up on a durable timer, and
    answers Queries about its state the whole time.
- type: text
  contents: |-
    # Three ways to talk to a Workflow

    **Signal**: send it data. It does not reply. Here, `approve`.

    **Query**: ask it for its state. It changes nothing. Here, `getStatus`.

    **Timer**: wait a set time, kept by the server. It still fires if every
    Worker is down.
tabs:
- title: Code Editor
  type: code
  hostname: workshop
  path: /root/workshop/go/lab2
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
  path: /root/workshop/go/solution2
- title: CLI
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
difficulty: basic
timelimit: 1800
enhanced_loading: null
---

# Human in the Loop

A money transfer withdraws, then waits for approval. It gets one of three
endings: approved, rejected, or out of time.

> [!NOTE]
> This lab has a sixth tab. The starter keeps the
> [button label="Terminal" background="#444CE7"](tab-2) busy while it waits, so
> send Signals and Queries from the
> [button label="CLI" background="#444CE7"](tab-5).

# The Code

Every TODO is in `workflow.go`, in the
[button label="Code Editor" background="#444CE7"](tab-0). The Activities,
models, Worker and starter are finished.

| TODO | What it does |
|------|--------------|
| 1, Query | `getStatus` returns the current `status` |
| 2, Signal | A goroutine receives one `bool` from the `approve` channel |
| 3, Timer | Wait up to 30 seconds for that decision, then deposit or refund |

The timer is a single call:

```go
gotDecision, err := workflow.AwaitWithTimeout(ctx, ApprovalTimeout, func() bool { return decided })
```

`gotDecision` is `false` when the 30 seconds ran out first.

> [!IMPORTANT]
> Delete each `// TODO` comment as you finish. `lab2/SOLUTION.md` has every
> snippet.

# Run 1: Approve It

Start the Worker in the [button label="Worker" background="#444CE7"](tab-1):

```bash,run
go run ./lab2/worker
```

Start a transfer in the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab2/starter
```

You have 30 seconds. In the [button label="CLI" background="#444CE7"](tab-5),
ask where it is:

```bash,run
temporal workflow query --workflow-id money-transfer-workflow --name getStatus
```

```bash,nocopy
Query result:
  QueryResult  "AWAITING_APPROVAL"
```

Approve it:

```bash,run
temporal workflow signal --workflow-id money-transfer-workflow --name approve --input true
```

The starter prints `Final status: COMPLETED`.

# Run 2: Stop the Worker and Let It Time Out

Start another transfer in the
[button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab2/starter
```

Now stop the Worker: `Ctrl+C` in the
[button label="Worker" background="#444CE7"](tab-1). Run the Query again from
the [button label="CLI" background="#444CE7"](tab-5):

```bash,run
temporal workflow query --workflow-id money-transfer-workflow --name getStatus
```

It hangs, then fails, because a Query is answered by a Worker and none is
running:

```bash,nocopy
Error: querying workflow failed: no poller seen for task queue recently, worker may be down
```

That takes about 30 seconds, about as long as the approval deadline. Nothing
is running your code, but the server keeps the timer anyway. Once 30 seconds
have passed, restart the Worker:

```bash,run
go run ./lab2/worker
```

The Worker picks up the fired timer straight away, refunds, and the starter
prints `Final status: TIMED_OUT`. Nothing was lost while no Worker was running.

# Read the History

In the [button label="Temporal UI" background="#444CE7"](tab-3), compare the
two runs of `money-transfer-workflow`:

- **Run 1**: `TimerStarted`, then `WorkflowExecutionSignaled`, then the deposit.
- **Run 2**: `TimerStarted`, then `TimerFired`, then the refund.

Queries never show up in the history. They read state and leave no record.

Click **Check** once you have one approved transfer and one that timed out.

# Key Takeaways

- A Signal pushes data into a running Workflow. It does not wait for a reply.
- A Query reads state and writes nothing to the history. It needs a Worker to answer.
- `AwaitWithTimeout` is a durable timer. The server keeps it, so it fires even when every Worker is down.
- A stopped Worker pauses progress but loses nothing. The history is the state.
