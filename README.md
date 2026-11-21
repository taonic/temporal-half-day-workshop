# Temporal Half-Day Workshop

Three hands-on labs for a four-hour Temporal workshop, in one repo per
language. Each lab is three short pages; each page adds one concept and ends
with something you run and watch. Each language also ships as its own
Instruqt track.

| Lab | Pages | You build |
|---|---|---|
| 1. Hello Temporal, with retries | 1a Workflow, Activity, Worker, client · 1b a flaky service · 1c a retry policy | A greeting that arrives although its service fails twice |
| 2. Human in the loop | 2a Signal · 2b Query · 2c durable timer | A money transfer that waits for approval, or gives up after 30 seconds |
| 3. Failure handling and testing | 3a non-retryable error · 3b Update · 3c time-skipping tests | A bad account parks the transfer until an Update fixes it, plus tests |

Nobody types or pastes code. Every page shows the code it adds, and one
command, `apply <region>`, puts that code in place. Then you run it, and try
one change of your own.

Every language implements the same contract: [SPEC.md](SPEC.md). Same pages,
region names, task queues, Workflow IDs, message names and status strings,
so the Instruqt checks in [`shared/instruqt/`](shared/instruqt/) and the
`apply` command in [`shared/steps.sh`](shared/steps.sh) work for all of them.

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

The step-by-step instructions are the Instruqt assignments, under
`<language>/instruqt/track/*/assignment.md`, one per page. Before each page,
`reset <page>` puts the lab folder at the start of that page. Then follow the
assignment: its `apply` commands fill in the code.

### Go

```bash
cd go
export PATH="$PWD/bin:$PATH"   # apply and reset

reset 1a                       # creates lab1 at the start of page 1a
apply compose-greeting         # ...and the rest of the page's apply commands
go run ./lab1/worker           # terminal 1
go run ./lab1/starter          # terminal 2

reset 1b                       # next page
```

`reset` and `apply` take the page explicitly too: `apply 1c activity-options`.
The finished code for every page is in [`go/steps/`](go/steps/).

## Instruqt

Each language folder has an `instruqt/` directory with its track and sandbox
preset. See [go/instruqt/README.md](go/instruqt/README.md) for how they are
built, tested locally and published.
