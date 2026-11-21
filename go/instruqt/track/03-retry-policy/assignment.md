---
slug: retry-policy
type: challenge
title: '1c: A Retry Policy That Works'
teaser: Give the Activity enough attempts, with backoff, and the greeting gets through.
tabs:
- title: Code Editor
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/go/lab1
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

# A Retry Policy That Works

The service still fails twice. Allow five attempts with backoff, and you
write no retry loop at all.

# The Retry Policy

- `InitialInterval`: wait 1 second before the first retry.
- `BackoffCoefficient`: double the wait after each failure, so 1s, then 2s.
- `MaximumAttempts`: give up after 5 attempts.

In `workflow.go`:

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

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
apply activity-options
```

# Run It

Restart the Worker in the [button label="Worker" background="#444CE7"](tab-1):
`Ctrl+C`, then:

```bash,run
go run ./lab1/worker
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab1/starter
```

The greeting takes about three seconds: attempt 1 fails, wait 1 second,
attempt 2 fails, wait 2 seconds, attempt 3 succeeds.

```bash,nocopy
Started greeting-workflow
Hello, Temporal!
```

# See It

In the [button label="Temporal UI" background="#444CE7"](tab-3), open
`greeting-workflow` and expand `ActivityTaskStarted`. It says `attempt: 3`,
and `lastFailure` holds the error from attempt 2.

While an Activity is retrying, the UI shows its attempt count and last
failure under **Pending Activities**. The CLI shows the history too:

```bash,run
temporal workflow show --workflow-id greeting-workflow
```

# Predict, Then Try

Set `BackoffCoefficient: 1.0`, restart the Worker, and run the starter. How
long does the greeting take now?

About two seconds: every retry waits the same 1 second.

Click **Check** once the greeting arrives on attempt 3.

# Key Takeaways for Lab 1

- A Workflow orchestrates. An Activity does the work that can fail.
- A Worker runs the code it was compiled with: restart it after a change.
- A retry policy goes in the Activity options, not in a loop you write.
- Retries back off: each wait is the last one times `BackoffCoefficient`.
- The history records the attempt that finished, with the last failure.
