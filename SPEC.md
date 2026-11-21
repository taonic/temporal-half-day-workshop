# Workshop Spec

The language-neutral contract for the workshop. Every language in this repo
implements exactly this. The Instruqt check and solve scripts in
[`shared/instruqt/`](shared/instruqt/) drive and inspect Workflows through the
`temporal` CLI only, and [`shared/steps.sh`](shared/steps.sh) applies code by
region name, so both work for every language **only if every page, region
name, Workflow name, payload and status string below matches exactly**.

When a port finds something here that its SDK cannot do, change this file
first, then every language.

## Common

| Item | Value |
|---|---|
| Server | `127.0.0.1:7233`, namespace `default` |
| Payload encoding | The SDK's default JSON converter |
| JSON field names | camelCase, exactly as listed below |

Workflow type names are registered **explicitly** where the SDK would
otherwise derive a different name (TypeScript function name, Java interface
name, .NET and Python class name). Activity type names are not part of the
contract: nothing outside the language's own tests refers to them.

## Pages

Three labs, nine pages. Each page is one Instruqt challenge (15 minutes) that
adds one concept, ends with a run the learner watches, and offers one
"predict, then try" experiment.

| Page | Challenge | Lab folder | Adds | Regions applied, in order |
|---|---|---|---|---|
| 1a | `first-workflow` | `lab1` | Activity, Workflow, Worker, client | `compose-greeting`; `activity-options` `execute-activity`; `register`; `start-workflow` |
| 1b | `flaky-service` | `lab1` | A failing service, `MaximumAttempts: 2` | `flaky-service`; `activity-options` |
| 1c | `retry-policy` | `lab1` | A full retry policy | `activity-options` |
| 2a | `signal` | `lab2` | Wait for the `approve` Signal | `approve-signal`; `wait-for-approval` |
| 2b | `query` | `lab2` | The `getStatus` Query | `get-status` |
| 2c | `durable-timer` | `lab2` | A 30 second deadline, Worker stopped | `wait-for-approval` |
| 3a | `non-retryable-error` | `lab3` | `InvalidAccount`, non-retryable | `invalid-account` |
| 3b | `update` | `lab3` | Park in `RETRYING`, `updateToAccount` | `update-to-account`; `deposit`; `wait-for-handlers` |
| 3c | `testing` | `lab3` | Two time-skipping tests | `test-timeout` `test-update` |

### Snapshots and regions

Each language folder has `steps/<snapshot>/`, a complete copy of one lab
folder: `N-start` before the lab's first page, and one per page after it
(`1a`, `1b`, ...). Learners work in `lab1`, `lab2`, `lab3`, which `reset`
creates from the snapshots. Page `Nx` starts from the snapshot before it:
`1a` from `1-start`, `1b` from `1a`. Lab 3's `3-start` is the finished Lab 2
transfer, plus Lab 3's constants and starter.

Code a page changes sits between markers in the language's line comment:

```
// region: activity-options
...
// endregion
```

- Regions do not nest. A region name may occur in several places and files
  (a Python Query handler and the state it reads); `apply` replaces them all.
- Every file whose imports change between pages has an `imports` region.
  `apply` syncs it silently with whatever it applies to that file.
- No region is ever empty. One that a later page fills holds the smallest
  valid code: a comment, a placeholder `return`, `pass`, or a skipped test.
- Every snapshot builds, and its tests (if any) pass or skip.
- Assignments show each applied region's exact code, then a run block
  `apply <region>...`. A drift test keeps the two in step.

### The commands

`shared/steps.sh` is language-neutral. Each language wraps it as
`<lang>/bin/apply` and `<lang>/bin/reset`, on PATH in the sandbox.

| Command | Does |
|---|---|
| `reset [page]` | Recreate the lab folder at the start of the page and make it the current page. Setup scripts run it |
| `apply [page] <region>...` | Copy the named regions, plus `imports`, from the page's snapshot into the lab folder. Refuses, changing nothing, if markers do not line up |

`apply` never restarts the Worker. It prints a reminder when it changed a
non-test file.

---

## Lab 1: Hello Temporal, with retries

| Item | Value |
|---|---|
| Task queue | `greeting-task-queue` |
| Workflow ID | `greeting-workflow` |
| Workflow type | `Greeting` |
| Input | one string, the name. The starter passes `"Temporal"` |
| Result | one string, `"Hello, <name>!"` |

**Activity `ComposeGreeting(name) -> string`** returns `"Hello, <name>!"`.
From page 1b it simulates a flaky greeting service, reading its own attempt
number from the Activity info:

- attempts 1 and 2: fails with a **retryable** application error, type
  `ServiceUnavailable`, message `greeting service unavailable (attempt N)`
- attempt 3 onwards: returns `"Hello, <name>!"`

**Activity options** (the `activity-options` region)

| Page | StartToClose | Retry policy |
|---|---|---|
| 1a | 5 seconds | none set |
| 1b | 5 seconds | maximum attempts 2 |
| 1c | 5 seconds | initial interval 1 second, backoff 2.0, maximum attempts 5 |

The starter starts the Workflow, waits for the result and prints it.

**Checks**, for runs started after the page began:

| Page | Passes when |
|---|---|
| 1a | A run completed with result `Hello, Temporal!` |
| 1b | A run failed, with an `ActivityTaskStarted` event at `attempt == 2` |
| 1c | A run completed with `Hello, Temporal!` and an `ActivityTaskStarted` at `attempt == 3` |

---

## Labs 2 and 3: Money transfer

| Item | Value |
|---|---|
| Task queue | `money-transfer-task-queue` |
| Workflow ID | `money-transfer-workflow` |
| Workflow type | `MoneyTransfer` |
| Input | `TransferRequest` |
| Result | the final `TransferStatus`, as a string |

**`TransferRequest`** (JSON):

```json
{"fromAccount": "account-123", "toAccount": "account-456", "amount": 100, "transferId": "transfer-1"}
```

**`TransferStatus`** values:

| Status | Meaning |
|---|---|
| `PENDING` | Started, withdrawal in progress |
| `AWAITING_APPROVAL` | Withdrawn, waiting for the `approve` Signal |
| `APPROVED` | Approved, deposit in progress |
| `COMPLETED` | Deposited. Terminal |
| `CANCELLED` | Rejected, refunded. Terminal |
| `TIMED_OUT` | No decision before the deadline, refunded. Terminal |
| `RETRYING` | Lab 3 only. Deposit hit an invalid account, waiting for `updateToAccount` |
| `FAILED` | Lab 3 only. A deposit failure that is not `InvalidAccount`. Terminal |

**Activities** (all take `account: string, amount: number`, return nothing):

| Activity | Lab 2 | Lab 3 |
|---|---|---|
| `Withdraw` | Logs, succeeds | Same |
| `Deposit` | Logs, succeeds | From page 3a: fails **non-retryable**, error type `InvalidAccount`, when `account == "invalid-account"`. Otherwise succeeds |
| `Refund` | Logs, succeeds | Same |

Activity options: 5 second StartToClose, retry 1s / 2.0 / 5 attempts.

**Message handlers**

| Name | Kind | Payload | Returns | From page |
|---|---|---|---|---|
| `approve` | Signal | `bool`, `true` approves | — | 2a |
| `getStatus` | Query | none | `TransferStatus` | 2b |
| `updateToAccount` | Update | `string`, the corrected account | `TransferStatus` | 3b |

**Approval deadline** (from page 2c): 30 seconds, a named constant.
Implemented as the SDK's wait-for-condition-with-timeout (`AwaitWithTimeout`,
`condition(fn, '30s')`, `wait_condition(timeout=)`,
`Workflow.await(Duration, ...)`, `WaitConditionAsync(fn, TimeSpan)`), which
records a durable timer in history. Page 2a waits with no deadline.

**Starters** start the transfer and **exit**, so every other command runs in
the same terminal. They replace a run that is still going (Workflow ID
conflict policy *terminate existing*). Learners follow a transfer with
`temporal workflow result --workflow-id money-transfer-workflow`.

### Lab 2: Human in the loop

```
PENDING --Withdraw--> AWAITING_APPROVAL --+-- approve(true)  --Deposit--> APPROVED -> COMPLETED
                                           +-- approve(false) --Refund---> CANCELLED
                                           +-- 30s timer      --Refund---> TIMED_OUT   (page 2c)
```

`2-start` withdraws and deposits with no wait. Activities never fail in Lab
2, so every run is deterministic. The starter transfers to `account-456`.
Page 2c stops the Worker while the timer runs, then restarts it.

**Checks**, for runs started after the page began:

| Page | Passes when |
|---|---|
| 2a | A run completed with `COMPLETED` and an `approve` Signal in its history |
| 2b | As 2a, and a `getStatus` Query on that run, answered by the learner's running Worker, returns `COMPLETED` |
| 2c | A run completed with `TIMED_OUT`, a `TimerFired` event and no Signal |

### Lab 3: Failure handling and testing

Starts from the finished Lab 2 transfer and changes the deposit step:

```
APPROVED --Deposit--> ok ---------------------------------------------> COMPLETED
                   +-> InvalidAccount (non-retryable) --> RETRYING                     (page 3b)
                   |      updateToAccount(acct) --> retry Deposit with acct
                   +-> any other error, after retries --> FAILED, Workflow fails
```

On page 3a there is no loop yet: the non-retryable error fails the Workflow.

**`updateToAccount(acct)`** (page 3b)

- **Validator** rejects, before anything reaches history, when the status is
  not `RETRYING`, or `acct` is empty, or `acct == "invalid-account"`.
- **Handler** stores `acct` as the deposit target, wakes the deposit loop,
  waits until the status leaves `RETRYING`, and returns the status. For a
  valid account that is `COMPLETED`.
- The Workflow waits for all handlers to finish before it returns.

The starter starts a transfer to `invalid-account`, sends `approve(true)`
after two seconds, and exits. The learner queries the status, tries an
invalid Update to see the validator reject it, then sends
`updateToAccount("account-456")` from the CLI.

**Tests** (page 3c): the test file ships with its setup done and both tests
skipped. Page 3c fills them:

1. **Timeout path** (`test-timeout`): mocked `Withdraw` and `Refund`, no
   Signal, time skipped past the deadline. Asserts the result is `TIMED_OUT`
   and `Refund` ran once.
2. **Invalid account fixed by Update** (`test-update`): mocked `Deposit` fails
   with `InvalidAccount` for `invalid-account` and succeeds for
   `account-456`. The test sends `approve(true)`, then
   `updateToAccount("account-456")`. Asserts the Update returned `COMPLETED`
   and the Workflow result is `COMPLETED`.

| Language | Test framework | Environment for test 2 |
|---|---|---|
| Go | `go.temporal.io/sdk/testsuite` | In-memory test environment (supports Updates) |
| TypeScript | to be decided in the port | |
| Python | to be decided in the port | |
| Java | to be decided in the port | |
| .NET | to be decided in the port | |

Test 1 is time-skipped in every language. Test 2 may use the local dev-server
test environment where the time-skipping server does not support Updates.

**Checks**, for runs started after the page began:

| Page | Passes when |
|---|---|
| 3a | A run failed with a non-retryable `ActivityTaskFailed` |
| 3b | A run completed with `COMPLETED`, a non-retryable `ActivityTaskFailed`, and an accepted and completed `updateToAccount` Update |
| 3c | Both tests run and pass: two passing test cases, none skipped |
