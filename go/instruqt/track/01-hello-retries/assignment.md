---
slug: hello-retries
id: ctao8fanukbb
type: challenge
title: 'Lab 1: Hello Temporal, with Retries'
teaser: Write a Workflow, an Activity and a Worker. The Activity fails twice, and
  the greeting still arrives.
notes:
- type: text
  contents: |-
    # Who retries when the greeting service is down?

    Usually you do: a loop, a sleep, a counter, and a bug in one of them.

    In Temporal the retry policy is part of the Activity call. The Activity
    in this lab fails on its first two attempts on purpose. You write no
    retry loop, and the greeting still arrives.
- type: text
  contents: |-
    # Already running in your sandbox

    A Temporal dev server on 127.0.0.1:7233, with the Web UI on port 8233.

    Go 1.24 is installed and the Temporal SDK is already compiled into the
    build cache, so `go run` starts in seconds.

    The Solution tab has the finished code. `SOLUTION.md` in the lab folder
    has every snippet ready to paste. Copying is fine.
tabs:
- id: l59xgxnzjmab
  title: Code Editor
  type: code
  hostname: workshop
  path: /root/workshop/go/lab1
- id: ffzzbogig5cm
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: k1txxmzacplv
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: uhs8i2isaxkh
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
- id: zsgkja3n3fwp
  title: Solution
  type: code
  hostname: workshop
  path: /root/workshop/go/solution1
difficulty: basic
timelimit: 1800
enhanced_loading: null
---

# Hello Temporal, with Retries

Four files, five TODOs, about 15 lines of Go. Then watch Temporal retry a
failing Activity for you.

> [!NOTE]
> **Sandbox Notes:**
> - [button label="Code Editor" background="#444CE7"](tab-0) edits the files in `lab1/`
> - [button label="Worker" background="#444CE7"](tab-1) runs the Worker
> - [button label="Terminal" background="#444CE7"](tab-2) starts Workflows
> - [button label="Temporal UI" background="#444CE7"](tab-3) is the Temporal Web UI
> - [button label="Solution" background="#444CE7"](tab-4) has the finished code
>
> The blue buttons are clickable. Instruqt saves editor changes for you.

# The Code

Open the [button label="Code Editor" background="#444CE7"](tab-0):

| File | What goes in it |
|------|-----------------|
| `activities.go` | `ComposeGreeting` returns `"Hello, <name>!"`. The code that fails attempts 1 and 2 is already there |
| `workflow.go` | Activity options with a retry policy, then execute `ComposeGreeting` and return its result |
| `worker/main.go` | Register `Greeting` and `ComposeGreeting` |
| `starter/main.go` | Start the `Greeting` Workflow with `"Temporal"` |

The retry policy is the heart of this lab:

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 5 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumAttempts:    5,
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
```

Attempt 1 fails, Temporal waits 1 second. Attempt 2 fails, it waits 2 seconds.
Attempt 3 succeeds.

> [!IMPORTANT]
> Delete each `// TODO` comment as you finish it. **Check** looks for leftover
> TODOs before anything else. Stuck? `lab1/SOLUTION.md` has every snippet.

# Start the Worker

Click the [button label="Worker" background="#444CE7"](tab-1):

```bash,run
go run ./lab1/worker
```

Leave it running. A Go Worker runs the code it was compiled with, so after
editing the Workflow or the Activity, stop it with `Ctrl+C` and start it again.

# Run It

Click the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab1/starter
```

The greeting takes about three seconds to arrive:

```bash,nocopy
Hello, Temporal!
```

Look back at the [button label="Worker" background="#444CE7"](tab-1). It logged
`Composing greeting` three times, with `attempt` 1, 2 and 3, and an error
after each of the first two.

# Read the History

Click the [button label="Temporal UI" background="#444CE7"](tab-3), then
`greeting-workflow`.

Open the **Event History** and expand `ActivityTaskStarted`. It says
`attempt: 3`, and `lastFailure` holds the error from attempt 2.

There is only **one** `ActivityTaskStarted`, not three. Temporal does not write
an event for every retry. While an Activity is retrying, the UI shows its
attempt count and last failure under **Pending Activities**. The history only
records the attempt that finished.

The CLI shows the same thing:

```bash,run
temporal workflow show --workflow-id greeting-workflow
```

Click **Check** when the greeting prints.

# Key Takeaways

- A Workflow orchestrates. An Activity does the work that can fail.
- A retry policy goes in the Activity options, not in a loop you write.
- Retries back off: `InitialInterval` × `BackoffCoefficient` for each attempt.
- The history records the attempt that finished, along with the last failure.
