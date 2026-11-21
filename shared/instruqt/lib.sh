#!/bin/bash
# ABOUTME: Language-neutral helpers for every track's lifecycle scripts.
#
# Sourced at runtime from the sandbox's clone of this repo:
#
#   . /root/workshop/shared/instruqt/lib.sh
#
# Everything here talks to Temporal through the `temporal` CLI only, so it
# works for every language as long as the language follows SPEC.md: same
# pages, task queues, Workflow IDs, Workflow types, message names, payloads
# and status strings.
#
# What stays in each track's own helpers: how to reset a page, run tests,
# and start and stop that language's Worker.

TEMPORAL_ARGS=(--address 127.0.0.1:7233 --namespace default)

GREETING_WF=greeting-workflow
GREETING_TQ=greeting-task-queue
TRANSFER_WF=money-transfer-workflow
TRANSFER_TQ=money-transfer-task-queue

fail() { echo "FAIL: $1"; exit 1; }
pass() { echo "PASS: $1"; }

# ---------------------------------------------------------------------------
# Page bookkeeping
# ---------------------------------------------------------------------------

# Pages reuse Workflow IDs, so "a transfer completed" cannot tell this page's
# run from the last one. Setup stamps the start time and the checks count
# only runs that started after it. Nanoseconds, because pages can follow each
# other within a second under `instruqt track test`.
stamp_start() { date -u +%Y-%m-%dT%H:%M:%S.%NZ > "/tmp/workshop-start-$1"; }
since()       { cat "/tmp/workshop-start-$1" 2>/dev/null || echo "1970-01-01T00:00:00Z"; }

# A run still waiting on a Signal or Update would confuse the next page.
# Terminate both IDs on every page boundary.
terminate_workshop_workflows() {
    local wf
    for wf in "$GREETING_WF" "$TRANSFER_WF"; do
        temporal workflow terminate "${TEMPORAL_ARGS[@]}" \
            --workflow-id "$wf" --reason "${1:-workshop reset}" >/dev/null 2>&1 || true
    done
}

# ---------------------------------------------------------------------------
# History inspection
# ---------------------------------------------------------------------------

# runs <workflow-id> <status> <since>: run IDs, newest first. Status is a
# visibility ExecutionStatus: Completed, Failed, Running, ...
runs() {
    temporal workflow list "${TEMPORAL_ARGS[@]}" -o json \
        --query "WorkflowId=\"$1\" AND ExecutionStatus=\"$2\" AND StartTime > \"$3\"" \
        2>/dev/null | jq -r '.[]?.execution.runId' 2>/dev/null
}

# history <workflow-id> <run-id>: the full event history as JSON.
history() {
    temporal workflow show "${TEMPORAL_ARGS[@]}" --workflow-id "$1" --run-id "$2" -o json 2>/dev/null
}

# result_of: decoded Workflow result from a history on stdin, as JSON.
result_of() {
    jq -r '.events[]
        | select(.eventType == "EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED")
        | .workflowExecutionCompletedEventAttributes.result.payloads[0].data
        | @base64d' 2>/dev/null
}

# count_events <event-type-suffix> [jq filter]: events in a history on stdin.
count_events() {
    jq --arg t "EVENT_TYPE_$1" "[.events[] | select(.eventType == \$t) ${2:+| select($2)}] | length" 2>/dev/null
}

signals() { count_events WORKFLOW_EXECUTION_SIGNALED "${1:+.workflowExecutionSignaledEventAttributes.signalName == \"$1\"}"; }
started_on_attempt() { count_events ACTIVITY_TASK_STARTED ".activityTaskStartedEventAttributes.attempt == $1"; }
non_retryable_failures() { count_events ACTIVITY_TASK_FAILED '.activityTaskFailedEventAttributes.failure.applicationFailureInfo.nonRetryable == true'; }

# ---------------------------------------------------------------------------
# Runtime checks, one per page. Each exits through fail() on the first
# missing piece. Page 3c (tests) is checked by the track's own helpers.
# ---------------------------------------------------------------------------

check_1a() {
    local run h
    for run in $(runs "$GREETING_WF" Completed "$1"); do
        h=$(history "$GREETING_WF" "$run")
        if [ "$(result_of <<<"$h")" = '"Hello, Temporal!"' ]; then
            pass "$GREETING_WF returned \"Hello, Temporal!\""
            return 0
        fi
        fail "$GREETING_WF returned $(result_of <<<"$h"), not \"Hello, Temporal!\". Apply compose-greeting, restart the Worker and run the starter again."
    done
    fail "No greeting yet on this page. Apply all four changes, start the Worker, then run the starter."
}

check_1b() {
    local run h
    for run in $(runs "$GREETING_WF" Failed "$1"); do
        h=$(history "$GREETING_WF" "$run")
        if [ "$(started_on_attempt 2 <<<"$h")" -ge 1 ]; then
            pass "$GREETING_WF failed after 2 attempts, as the retry policy allows"
            return 0
        fi
    done
    if [ -n "$(runs "$GREETING_WF" Completed "$1")" ]; then
        fail "The greeting still arrives. Apply flaky-service and activity-options, restart the Worker, then run the starter again."
    fi
    fail "No failed greeting yet on this page. Apply both changes, restart the Worker, then run the starter."
}

check_1c() {
    local run h
    for run in $(runs "$GREETING_WF" Completed "$1"); do
        h=$(history "$GREETING_WF" "$run")
        [ "$(result_of <<<"$h")" = '"Hello, Temporal!"' ] || continue
        if [ "$(started_on_attempt 3 <<<"$h")" -ge 1 ]; then
            pass "ComposeGreeting succeeded on attempt 3, after two automatic retries"
            return 0
        fi
    done
    fail "No greeting on attempt 3 yet. Apply activity-options, restart the Worker, then run the starter."
}

# approved_run <since>: the newest transfer completed through an approve Signal.
approved_run() {
    local run h
    for run in $(runs "$TRANSFER_WF" Completed "$1"); do
        h=$(history "$TRANSFER_WF" "$run")
        if [ "$(result_of <<<"$h")" = '"COMPLETED"' ] && [ "$(signals approve <<<"$h")" -ge 1 ]; then
            echo "$run"
            return 0
        fi
    done
    return 1
}

check_2a() {
    approved_run "$1" >/dev/null ||
        fail "No transfer has completed through the approve Signal on this page yet. Run the starter, then send: temporal workflow signal --workflow-id $TRANSFER_WF --name approve --input true"
    pass "A transfer waited for the approve Signal and completed"
}

check_2b() {
    local run answer
    run=$(approved_run "$1") ||
        fail "No approved transfer on this page yet. Run the starter, query it, then approve it."
    pass "A transfer completed through the approve Signal"
    # Queries leave no trace in history, so ask the finished run directly.
    # A Worker answers by replaying the history, so one must be running.
    answer=$(timeout 15 temporal workflow query "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" \
        --run-id "$run" --name getStatus -o json 2>/dev/null | jq -r '.queryResult[0] // empty' 2>/dev/null)
    [ -n "$answer" ] || fail "The Worker did not answer getStatus. Is the Worker running the code with get-status applied?"
    [ "$answer" = COMPLETED ] || fail "getStatus answered $answer for a completed transfer. It should answer COMPLETED."
    pass "getStatus answers COMPLETED for the finished transfer"
}

check_2c() {
    local run h
    for run in $(runs "$TRANSFER_WF" Completed "$1"); do
        h=$(history "$TRANSFER_WF" "$run")
        if [ "$(result_of <<<"$h")" = '"TIMED_OUT"' ] &&
            [ "$(count_events TIMER_FIRED <<<"$h")" -ge 1 ] &&
            [ "$(signals <<<"$h")" -eq 0 ]; then
            pass "A transfer timed out on the durable timer and was refunded"
            return 0
        fi
    done
    fail "No transfer has timed out on this page yet. Run the starter, send nothing, and wait 30 seconds."
}

check_3a() {
    local run h
    for run in $(runs "$TRANSFER_WF" Failed "$1"); do
        h=$(history "$TRANSFER_WF" "$run")
        if [ "$(non_retryable_failures <<<"$h")" -ge 1 ]; then
            pass "Deposit failed once with a non-retryable error, and the Workflow failed"
            return 0
        fi
    done
    if [ -n "$(runs "$TRANSFER_WF" Completed "$1")" ]; then
        fail "The transfer to invalid-account completed. Apply invalid-account, restart the Worker, then run the starter again."
    fi
    fail "No failed transfer on this page yet. Apply invalid-account, restart the Worker, run the starter, and wait about 3 seconds."
}

check_3b() {
    local run h
    for run in $(runs "$TRANSFER_WF" Completed "$1"); do
        h=$(history "$TRANSFER_WF" "$run")
        [ "$(result_of <<<"$h")" = '"COMPLETED"' ] || continue
        [ "$(non_retryable_failures <<<"$h")" -ge 1 ] || continue
        [ "$(count_events WORKFLOW_EXECUTION_UPDATE_ACCEPTED '.workflowExecutionUpdateAcceptedEventAttributes.acceptedRequest.input.name == "updateToAccount"' <<<"$h")" -ge 1 ] || continue
        [ "$(count_events WORKFLOW_EXECUTION_UPDATE_COMPLETED <<<"$h")" -ge 1 ] || continue
        pass "Deposit failed with a non-retryable error"
        pass "updateToAccount was accepted and completed"
        pass "The transfer completed"
        return 0
    done
    fail "No transfer recovered through updateToAccount on this page yet. Run the starter, wait for RETRYING, then send the Update with account-456."
}

# ---------------------------------------------------------------------------
# Drivers for solve scripts. The track's solve script puts the page's
# finished code in place and starts its Worker; these then play the
# learner's part from the CLI.
# ---------------------------------------------------------------------------

wait_for_status() {
    local want=$1 status
    for _ in $(seq 1 30); do
        status=$(temporal workflow query "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" \
            --name getStatus -o json 2>/dev/null | jq -r '.queryResult[0] // empty' 2>/dev/null)
        [ "$status" = "$want" ] && return 0
        sleep 1
    done
    echo "Status never reached $want (last: ${status:-unknown})" >&2
    return 1
}

# Page 2a has no Query yet, so wait on the history instead: the transfer is
# waiting once Withdraw has completed.
wait_for_withdraw() {
    for _ in $(seq 1 30); do
        temporal workflow show "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" -o json 2>/dev/null |
            jq -e '[.events[] | select(.eventType == "EVENT_TYPE_ACTIVITY_TASK_COMPLETED")] | length >= 1' >/dev/null 2>&1 && return 0
        sleep 1
    done
    return 1
}

start_transfer() {
    temporal workflow start "${TEMPORAL_ARGS[@]}" \
        --type MoneyTransfer --task-queue "$TRANSFER_TQ" --workflow-id "$TRANSFER_WF" \
        --id-conflict-policy TerminateExisting \
        --input "{\"fromAccount\":\"account-123\",\"toAccount\":\"$1\",\"amount\":100,\"transferId\":\"transfer-solve\"}"
}

result() { timeout "${1:-60}" temporal workflow result "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF"; }

drive_greeting() {
    timeout 60 temporal workflow execute "${TEMPORAL_ARGS[@]}" \
        --type Greeting --task-queue "$GREETING_TQ" --workflow-id "$GREETING_WF" \
        --input '"Temporal"'
}

drive_1a() { drive_greeting; }
drive_1b() { drive_greeting || true; }   # fails on purpose
drive_1c() { drive_greeting; }

drive_2a() {
    start_transfer account-456
    wait_for_withdraw
    temporal workflow signal "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" --name approve --input true
    result
}

drive_2b() {
    start_transfer account-456
    wait_for_status AWAITING_APPROVAL
    temporal workflow signal "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" --name approve --input true
    result
}

drive_2c() {
    # Send nothing and let the 30 second timer fire.
    start_transfer account-456
    wait_for_status AWAITING_APPROVAL
    result 90
}

drive_3a() {
    start_transfer invalid-account
    wait_for_status AWAITING_APPROVAL
    temporal workflow signal "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" --name approve --input true
    result || true   # fails on purpose
}

drive_3b() {
    start_transfer invalid-account
    wait_for_status AWAITING_APPROVAL
    temporal workflow signal "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" --name approve --input true
    wait_for_status RETRYING
    timeout 60 temporal workflow update execute "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" \
        --name updateToAccount --input '"account-456"'
    result
}
