---
slug: query
type: challenge
title: '2b: Ask It With a Query'
teaser: Ask the waiting transfer where it is, without changing anything.
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

# Ask It With a Query

A waiting transfer is a black box. Give it a `getStatus` Query, so anyone
can ask where it is.

# The Query Handler

A Query handler is a function the Workflow registers by name. It reads
state and returns it. It must not change anything or block. `status` is
already kept up to date through the transfer.

```go
if err := workflow.SetQueryHandler(ctx, QueryGetStatus, func() (TransferStatus, error) {
    return status, nil
}); err != nil {
    return "", err
}
```

In the [button label="Terminal" background="#444CE7"](tab-2):

```bash,run
apply get-status
```

# Run It

Restart the Worker in the [button label="Worker" background="#444CE7"](tab-1):
`Ctrl+C`, then:

```bash,run
go run ./lab2/worker
```

In the [button label="Terminal" background="#444CE7"](tab-2), start a
transfer and ask where it is:

```bash,run
go run ./lab2/starter
```

```bash,run
temporal workflow query --workflow-id money-transfer-workflow --name getStatus
```

```bash,nocopy
Query result:
  QueryResult  "AWAITING_APPROVAL"
```

Approve it, then ask again:

```bash,run
temporal workflow signal --workflow-id money-transfer-workflow --name approve --input true
```

```bash,run
temporal workflow query --workflow-id money-transfer-workflow --name getStatus
```

```bash,nocopy
Query result:
  QueryResult  "COMPLETED"
```

A finished Workflow still answers. The Worker replays its history to rebuild
`status`, then runs the handler.

# See It

In the [button label="Temporal UI" background="#444CE7"](tab-3), look for
your Queries in the history. There are none: a Query reads state and leaves
no record.

# Predict, Then Try

Ask for a Query that does not exist:

```bash,run
temporal workflow query --workflow-id money-transfer-workflow --name getBalance
```

The error lists the Queries the Workflow does know.

Leave the Worker running and click **Check**: it asks your Worker for the
finished transfer's status.
