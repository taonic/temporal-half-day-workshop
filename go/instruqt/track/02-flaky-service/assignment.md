---
slug: flaky-service
id: tqigto8xa3vp
type: challenge
title: '1b: A Flaky Service'
teaser: The greeting service starts failing. See what a strict retry policy does.
notes:
- type: text
  contents: |-
    # Who retries when the greeting service is down?

    Usually you do: a loop, a sleep, a counter, and a bug in one of them.

    In Temporal the retry policy is part of the Activity call. This page
    makes the greeting service fail, and gives the Activity too few
    attempts to get through.
tabs:
- id: ypbhqorrelmg
  title: Code Editor
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/go/lab1
  port: 8080
- id: wprzjya7q276
  title: Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: ymumvyc0vjgp
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/go
- id: zx1vg7nn53ad
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
difficulty: basic
timelimit: 900
enhanced_loading: null
---

# A Flaky Service

Real services fail. Make the greeting service fail its first two calls, and
allow only two attempts.

# 1. Break the Service

`ComposeGreeting` now fails on attempts 1 and 2. The attempt number comes
from Temporal, which counts for you. In `activities.go`:

```go
attempt := activity.GetInfo(ctx).Attempt
if attempt < 3 {
    return "", temporal.NewApplicationError(
        fmt.Sprintf("greeting service unavailable (attempt %d)", attempt),
        "ServiceUnavailable",
    )
}
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
apply flaky-service
```

# 2. Allow Two Attempts

A `RetryPolicy` on the Activity options says how Temporal retries a failed
Activity. `MaximumAttempts: 2` means: try, retry once, then give up. In
`workflow.go`:

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 5 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        MaximumAttempts: 2,
    },
}
ctx = workflow.WithActivityOptions(ctx, ao)
```

```bash,run
apply activity-options
```

# Run It

The Worker still runs the old code. In the
[button label="Worker" background="#444CE7"](tab-1), stop it with `Ctrl+C`,
then start it again:

```bash,run
go run ./lab1/worker
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab1/starter
```

After about a second the Workflow fails:

```bash,nocopy
Started greeting-workflow
... Workflow failed workflow execution error (type: Greeting, ...): activity error (type: ComposeGreeting, ...): greeting service unavailable (attempt 2) (type: ServiceUnavailable, retryable: true)
exit status 1
```

The Worker logged an `Activity error` for attempt 1 and attempt 2.

# See It

In the [button label="Temporal UI" background="#444CE7"](tab-3), open the
failed `greeting-workflow`. Expand `ActivityTaskStarted`: `attempt` is 2.
Then `ActivityTaskFailed` holds the last error, and the Workflow failed.

There is only one `ActivityTaskStarted`, not two. Temporal records the
attempt that finished, not every retry.

# Predict, Then Try

In `workflow.go`, set `MaximumAttempts: 0`. Restart the Worker and run the
starter. What does 0 attempts mean?

It succeeds, on attempt 3. Zero means no limit, which is also Temporal's
default when you set no retry policy at all: retry forever, waiting 1
second, then 2, then 4, and so on, up to 100 seconds between attempts.

Click **Check** once you have seen a greeting fail after two attempts.
