---
slug: first-workflow
type: challenge
title: '1a: Your First Workflow'
teaser: An Activity, a Workflow, a Worker and a client. Four small pieces, one greeting.
notes:
- type: text
  contents: |-
    # The four pieces of every Temporal app

    **Activity**: ordinary code that does one job, such as calling a service.

    **Workflow**: decides which Activities run, in what order. Temporal
    records every step, so a Workflow survives crashes and restarts.

    **Worker**: your process. It runs your Workflows and Activities.

    **Client**: starts Workflows and talks to them.
- type: text
  contents: |-
    # How these labs work

    Every page shows the code it adds. A **run** button under the code puts
    it into the editor for you: no typing, no copy-paste.

    Then you run it, watch what happens, and try one change of your own.
    Each page starts from working code, so experiment freely.
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

# Your First Workflow

Build a Workflow that asks an Activity for a greeting, then run it.

> [!NOTE]
> **How to use this page**
> - Read each piece of code, then click **run** on the `apply` command under
>   it. The code lands in the [button label="Code Editor" background="#444CE7"](tab-0).
> - Run `apply` in the [button label="Terminal" background="#444CE7"](tab-2).
>   The Worker runs in the [button label="Worker" background="#444CE7"](tab-1).
> - The [button label="Temporal UI" background="#444CE7"](tab-3) shows every Workflow and its history.

# 1. The Activity

An Activity is plain Go. It does the work that talks to the outside world,
where things fail. This one builds the greeting, in `activities.go`:

```go
activity.GetLogger(ctx).Info("Composing greeting", "name", name, "attempt", activity.GetInfo(ctx).Attempt)
return fmt.Sprintf("Hello, %s!", name), nil
```

Open the [button label="Terminal" background="#444CE7"](tab-2) and apply it:

```bash,run
apply compose-greeting
```

# 2. The Workflow

The Workflow does not call `ComposeGreeting` directly. It asks Temporal to
run it, with options. `StartToCloseTimeout` says how long one attempt may
take. This goes in `workflow.go`:

```go
ao := workflow.ActivityOptions{
    StartToCloseTimeout: 5 * time.Second,
}
ctx = workflow.WithActivityOptions(ctx, ao)
```

```go
var greeting string
err := workflow.ExecuteActivity(ctx, ComposeGreeting, name).Get(ctx, &greeting)
if err != nil {
    return "", err
}
return greeting, nil
```

```bash,run
apply activity-options execute-activity
```

`.Get` waits for the Activity's result. While it waits, the Workflow uses no
thread: Temporal wakes it when the result is in.

# 3. The Worker

The Worker polls a **Task Queue** for work, and runs the Workflows and
Activities registered with it. In `worker/main.go`:

```go
w.RegisterWorkflow(lab1.Greeting)
w.RegisterActivity(lab1.ComposeGreeting)
```

```bash,run
apply register
```

# 4. The Client

The starter asks the Temporal server to start a `Greeting` Workflow on the
Worker's Task Queue, then waits for the result. In `starter/main.go`:

```go
options := client.StartWorkflowOptions{
    ID:        "greeting-workflow",
    TaskQueue: lab1.TaskQueue,
}
we, err := c.ExecuteWorkflow(context.Background(), options, lab1.Greeting, "Temporal")
if err != nil {
    log.Fatalln("Unable to start workflow", err)
}
fmt.Println("Started", we.GetID())

var greeting string
if err := we.Get(context.Background(), &greeting); err != nil {
    log.Fatalln("Workflow failed", err)
}
fmt.Println(greeting)
```

```bash,run
apply start-workflow
```

# Run It

Start the Worker in the [button label="Worker" background="#444CE7"](tab-1)
and leave it running:

```bash,run
go run ./lab1/worker
```

Start the Workflow from the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
go run ./lab1/starter
```

```bash,nocopy
Started greeting-workflow
Hello, Temporal!
```

# See It

In the [button label="Temporal UI" background="#444CE7"](tab-3), open
`greeting-workflow`. The **Event History** is the Workflow's durable record:
`WorkflowExecutionStarted`, then `ActivityTaskScheduled`,
`ActivityTaskStarted` and `ActivityTaskCompleted` with the greeting, then
`WorkflowExecutionCompleted`.

# Predict, Then Try

In the [button label="Code Editor" background="#444CE7"](tab-0), change
`"Hello, %s!"` to `"Hi, %s!"` in `activities.go`. Run the starter again
**without** restarting the Worker. Which greeting prints?

It still says Hello: the Worker runs the code it was compiled with. Stop it
with `Ctrl+C`, start it again, and run the starter once more.

Click **Check** once a greeting has printed.
