# Workshop Spec

The language-neutral contract for the three labs. Every language in this repo
implements exactly this. The Instruqt check and solve scripts in
[`shared/instruqt/`](shared/instruqt/) drive and inspect Workflows through the
`temporal` CLI only, so they work for every language **only if every name,
payload and status string below matches exactly**.

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

Each lab has starter code with `TODO` comments in `labN/` and the finished
code in `solutionN/`. A lab is done when no `TODO` remains in the files that
lab lists.

---

## Lab 1: Hello Temporal, with retries

| Item | Value |
|---|---|
| Task queue | `greeting-task-queue` |
| Workflow ID | `greeting-workflow` |
| Workflow type | `Greeting` |
| Input | one string, the name. The starter passes `"Temporal"` |
| Result | one string, `"Hello, <name>!"` |

**Activity `ComposeGreeting(name) -> string`** simulates a flaky greeting
service. It reads its own attempt number from the Activity info and:

- attempts 1 and 2: fails with a **retryable** application error, type
  `ServiceUnavailable`, message `greeting service unavailable (attempt N)`
- attempt 3 onwards: returns `"Hello, <name>!"`

The failure logic is shipped in the starter code. The learner writes the
success path.

**Activity options**

| Option | Value |
|---|---|
| StartToClose timeout | 5 seconds |
| Retry: initial interval | 1 second |
| Retry: backoff coefficient | 2.0 |
| Retry: maximum attempts | 5 |

**Learner TODOs**: the Activity's success path; the Activity options and retry
policy; executing the Activity from the Workflow; registering the Workflow and
Activity on the Worker; starting the Workflow from the starter.

**Check passes when**, for a run started after the challenge began: it
completed, returned `Hello, Temporal!`, and its history has an
`ActivityTaskStarted` event with `attempt == 3`.

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
| `Deposit` | Logs, succeeds | Fails **non-retryable**, error type `InvalidAccount`, when `account == "invalid-account"`. Otherwise succeeds |
| `Refund` | Logs, succeeds | Same |

Activity options are Lab 1's: 5 second StartToClose, retry 1s / 2.0 / 5
attempts.

**Message handlers**

| Name | Kind | Payload | Returns | Lab |
|---|---|---|---|---|
| `getStatus` | Query | none | `TransferStatus` | 2, 3 |
| `approve` | Signal | `bool`, `true` approves | — | 2, 3 |
| `updateToAccount` | Update | `string`, the corrected account | `TransferStatus` | 3 |

**Approval deadline**: 30 seconds, a named constant. Implemented as the SDK's
wait-for-condition-with-timeout (`AwaitWithTimeout`, `condition(fn, '30s')`,
`wait_condition(timeout=)`, `Workflow.await(Duration, ...)`,
`WaitConditionAsync(fn, TimeSpan)`), which records a durable timer in history.

### Lab 2: Human in the loop

```
PENDING --Withdraw--> AWAITING_APPROVAL --+-- approve(true)  --Deposit--> APPROVED -> COMPLETED
                                           +-- approve(false) --Refund---> CANCELLED
                                           +-- 30s timer      --Refund---> TIMED_OUT
```

Activities never fail in Lab 2, so every run is deterministic.

The starter starts the Workflow and waits for its result. It sends nothing:
the learner sends `approve` and queries `getStatus` from the CLI. The
assignment includes stopping the Worker during the wait and restarting it.

**Learner TODOs**: the `getStatus` Query handler; the `approve` Signal
handler; the wait-with-timeout and the three branches with their status
updates.

**Check passes when**, among runs started after the challenge began, there is
at least one that completed with result `COMPLETED` and an `approve` Signal in
its history, and at least one that completed with result `TIMED_OUT`, a
`TimerFired` event and no Signal.

### Lab 3: Failure handling and testing

Starts from the Lab 2 solution and changes the deposit step:

```
APPROVED --Deposit--> ok ---------------------------------------------> COMPLETED
                   +-> InvalidAccount (non-retryable) --> RETRYING
                   |      updateToAccount(acct) --> retry Deposit with acct
                   +-> any other error, after retries --> FAILED, Workflow fails
```

**`updateToAccount(acct)`**

- **Validator** rejects, before anything reaches history, when the status is
  not `RETRYING`, or `acct` is empty, or `acct == "invalid-account"`.
- **Handler** stores `acct` as the deposit target, wakes the deposit loop,
  waits until the status leaves `RETRYING`, and returns the status. For a
  valid account that is `COMPLETED`.
- The Workflow waits for all handlers to finish before it returns.

The starter starts a transfer to `invalid-account`, sends `approve(true)`
after two seconds, and waits for the result. The learner queries the status,
tries an invalid Update to see the validator reject it, then sends
`updateToAccount("account-456")` from the CLI.

**Tests**: the test file ships with its setup done. The learner writes:

1. **Timeout path**: mocked `Withdraw` and `Refund`, no Signal, time skipped
   past the deadline. Asserts the result is `TIMED_OUT` and `Refund` ran once.
2. **Invalid account fixed by Update**: mocked `Deposit` fails with
   `InvalidAccount` for `invalid-account` and succeeds for `account-456`.
   The test sends `approve(true)`, then `updateToAccount("account-456")`.
   Asserts the Update returned `COMPLETED` and the Workflow result is
   `COMPLETED`.

| Language | Test framework | Environment for test 2 |
|---|---|---|
| Go | `go.temporal.io/sdk/testsuite` | In-memory test environment (supports Updates) |
| TypeScript | to be decided in the port | |
| Python | to be decided in the port | |
| Java | to be decided in the port | |
| .NET | to be decided in the port | |

Test 1 is time-skipped in every language. Test 2 may use the local dev-server
test environment where the time-skipping server does not support Updates.

**Learner TODOs**: the `InvalidAccount` non-retryable error in `Deposit`; the
deposit loop that parks in `RETRYING`; the `updateToAccount` handler and
validator; waiting for handlers before returning; the two tests.

**Check passes when** the tests pass with at least two test cases, and among
runs started after the challenge began there is one that completed with
result `COMPLETED`, a non-retryable `ActivityTaskFailed`, and an accepted and
completed `updateToAccount` Update in its history.
