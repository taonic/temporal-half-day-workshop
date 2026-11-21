# Temporal Half-Day Workshop

Three 30-minute hands-on labs for a four-hour Temporal workshop, in one repo
per language. Each language also ships as its own Instruqt track.

| Lab | You build | Concepts |
|---|---|---|
| 1. Hello Temporal, with retries | A greeting Workflow whose Activity fails twice before it succeeds | Workflow, Activity, Worker, Task Queue, retry policy |
| 2. Human in the loop | A money transfer that waits for approval or a 30-second deadline | Signal, Query, durable timer, Worker restart |
| 3. Failure handling and testing | A bad account parks the transfer until an Update fixes it, plus tests | Non-retryable errors, Update with validator, time-skipping tests, mocks |

Every language implements the same contract: [SPEC.md](SPEC.md). Same task
queues, Workflow IDs, message names and status strings, so the Instruqt checks
in [`shared/instruqt/`](shared/instruqt/) work for all of them.

| Language | Folder | Status |
|---|---|---|
| Go | [`go/`](go/) | Ready |
| TypeScript | `typescript/` | Planned |
| Python | `python/` | Planned |
| Java | `java/` | Planned |
| .NET | `dotnet/` | Planned |

## Run the labs on your own machine

Prerequisites: the [Temporal CLI](https://docs.temporal.io/cli#install) and
the language's toolchain (Go 1.24+ for `go/`).

Start a dev server and leave it running. The Web UI is on
<http://localhost:8233>.

```bash
temporal server start-dev
```

Each lab folder has starter code with `TODO` comments and a `SOLUTION.md` with
snippets to paste. The finished code is in the matching `solutionN/` folder.
The step-by-step instructions are the Instruqt assignments, under
`<language>/instruqt/track/*/assignment.md`.

### Go

```bash
cd go

# Lab 1
go run ./lab1/worker        # terminal 1
go run ./lab1/starter       # terminal 2

# Lab 2: approve within 30 seconds, or let it time out
go run ./lab2/worker
go run ./lab2/starter
temporal workflow query  --workflow-id money-transfer-workflow --name getStatus
temporal workflow signal --workflow-id money-transfer-workflow --name approve --input true

# Lab 3: the starter deposits to invalid-account; fix it with an Update
go run ./lab3/worker
go run ./lab3/starter
temporal workflow update execute --workflow-id money-transfer-workflow \
  --name updateToAccount --input '"account-456"'
go test -v ./lab3
```

Swap `labN` for `solutionN` to run the finished code.

## Instruqt

Each language folder has an `instruqt/` directory with its track and sandbox
preset. See [go/instruqt/README.md](go/instruqt/README.md) for how they are
built, tested locally and published.
