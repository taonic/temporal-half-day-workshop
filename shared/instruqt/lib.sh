#!/bin/bash
# ABOUTME: Language-neutral helpers for every track's lifecycle scripts.
#
# Sourced at runtime from the sandbox's clone of this repo:
#
#   . /root/workshop/shared/instruqt/lib.sh
#
# Everything here talks to Temporal through the `temporal` CLI only, so it
# works for every language as long as the language follows SPEC.md: same
# task queues, Workflow IDs, Workflow types, message names, payloads and
# status strings.
#
# What stays in each track's own scripts: the TODO file list, the build and
# test commands, and how to start and stop that language's Worker.

TEMPORAL_ARGS=(--address 127.0.0.1:7233 --namespace default)

GREETING_WF=greeting-workflow
GREETING_TQ=greeting-task-queue
TRANSFER_WF=money-transfer-workflow
TRANSFER_TQ=money-transfer-task-queue

fail() { echo "FAIL: $1"; exit 1; }
pass() { echo "PASS: $1"; }

# ---------------------------------------------------------------------------
# Challenge bookkeeping
# ---------------------------------------------------------------------------

# Labs 2 and 3 reuse one Workflow ID, so "a transfer completed" cannot tell
# this challenge's run from the last one. Setup stamps the start time and the
# checks count only runs that started after it.
stamp_start() { date -u +%Y-%m-%dT%H:%M:%SZ > "/tmp/workshop-start-$1"; }
since()       { cat "/tmp/workshop-start-$1" 2>/dev/null || echo "1970-01-01T00:00:00Z"; }

# A run still waiting on a Signal or Update would swallow the next starter:
# starting a Workflow ID that is already running attaches to the existing run
# in several SDKs. Terminate both IDs on every challenge boundary.
terminate_workshop_workflows() {
    local wf
    for wf in "$GREETING_WF" "$TRANSFER_WF"; do
        temporal workflow terminate "${TEMPORAL_ARGS[@]}" \
            --workflow-id "$wf" --reason "${1:-workshop reset}" >/dev/null 2>&1 || true
    done
}

# require_no_todos <dir> <file>...
require_no_todos() {
    local dir=$1 f
    shift
    for f in "$@"; do
        [ -f "$dir/$f" ] || fail "$dir/$f is missing."
        if grep -q "TODO" "$dir/$f"; then
            fail "$dir/$f still has TODO comments. Implement them, delete the comments, then check again."
        fi
    done
    pass "All TODOs implemented"
}

# ---------------------------------------------------------------------------
# History inspection
# ---------------------------------------------------------------------------

# completed_runs <workflow-id> <since>: run IDs, newest first.
completed_runs() {
    temporal workflow list "${TEMPORAL_ARGS[@]}" -o json \
        --query "WorkflowId=\"$1\" AND ExecutionStatus=\"Completed\" AND StartTime > \"$2\"" \
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

# ---------------------------------------------------------------------------
# Runtime checks, one per lab. Each runs after the track's own TODO and build
# gates, and exits through fail() on the first missing piece.
# ---------------------------------------------------------------------------

check_lab1_runtime() {
    local since=$1 runs run h
    runs=$(completed_runs "$GREETING_WF" "$since")
    [ -n "$runs" ] || fail "No completed $GREETING_WF yet. Start the Worker, run the starter, and wait for the greeting to print."
    pass "$GREETING_WF completed"

    for run in $runs; do
        h=$(history "$GREETING_WF" "$run")
        [ "$(result_of <<<"$h")" = '"Hello, Temporal!"' ] || continue
        if [ "$(count_events ACTIVITY_TASK_STARTED '.activityTaskStartedEventAttributes.attempt == 3' <<<"$h")" -ge 1 ]; then
            pass "Workflow returned \"Hello, Temporal!\""
            pass "ComposeGreeting succeeded on attempt 3, after two automatic retries"
            return 0
        fi
        fail "The greeting arrived without retries. ComposeGreeting must fail on attempts 1 and 2. Did the retry policy or the failure code change?"
    done
    fail "$GREETING_WF completed but did not return \"Hello, Temporal!\". Check the format string in ComposeGreeting."
}

check_lab2_runtime() {
    local since=$1 runs run h result approved=false timed_out=false
    runs=$(completed_runs "$TRANSFER_WF" "$since")
    [ -n "$runs" ] || fail "No completed transfer yet. Start the Worker, run the starter, then send the approve Signal."

    for run in $runs; do
        h=$(history "$TRANSFER_WF" "$run")
        result=$(result_of <<<"$h")
        if [ "$result" = '"COMPLETED"' ] &&
            [ "$(count_events WORKFLOW_EXECUTION_SIGNALED '.workflowExecutionSignaledEventAttributes.signalName == "approve"' <<<"$h")" -ge 1 ]; then
            approved=true
        fi
        if [ "$result" = '"TIMED_OUT"' ] &&
            [ "$(count_events TIMER_FIRED <<<"$h")" -ge 1 ] &&
            [ "$(count_events WORKFLOW_EXECUTION_SIGNALED <<<"$h")" -eq 0 ]; then
            timed_out=true
        fi
    done

    $approved || fail "No transfer has completed through the approve Signal yet. Run the starter and send: temporal workflow signal --workflow-id $TRANSFER_WF --name approve --input true"
    pass "A transfer completed after an approve Signal"
    $timed_out || fail "No transfer has timed out yet. Run the starter again and send nothing for 30 seconds."
    pass "A transfer timed out on the durable timer and was refunded"
}

check_lab3_runtime() {
    local since=$1 runs run h
    runs=$(completed_runs "$TRANSFER_WF" "$since")
    [ -n "$runs" ] || fail "No completed transfer yet. Run the Worker and the starter, then fix the account with the updateToAccount Update."

    for run in $runs; do
        h=$(history "$TRANSFER_WF" "$run")
        [ "$(result_of <<<"$h")" = '"COMPLETED"' ] || continue
        [ "$(count_events ACTIVITY_TASK_FAILED '.activityTaskFailedEventAttributes.failure.applicationFailureInfo.nonRetryable == true' <<<"$h")" -ge 1 ] || continue
        [ "$(count_events WORKFLOW_EXECUTION_UPDATE_ACCEPTED '.workflowExecutionUpdateAcceptedEventAttributes.acceptedRequest.input.name == "updateToAccount"' <<<"$h")" -ge 1 ] || continue
        [ "$(count_events WORKFLOW_EXECUTION_UPDATE_COMPLETED <<<"$h")" -ge 1 ] || continue
        pass "Deposit failed with a non-retryable error"
        pass "updateToAccount was accepted and completed"
        pass "The transfer completed"
        return 0
    done
    fail "No completed transfer recovered through updateToAccount yet. Run the starter (it deposits to invalid-account), wait for RETRYING, then send the Update."
}

# ---------------------------------------------------------------------------
# Drivers for solve scripts. The track's solve script copies the solution in
# and starts its Worker; these then play the learner's part from the CLI.
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

start_transfer() {
    temporal workflow start "${TEMPORAL_ARGS[@]}" \
        --type MoneyTransfer --task-queue "$TRANSFER_TQ" --workflow-id "$TRANSFER_WF" \
        --input "{\"fromAccount\":\"account-123\",\"toAccount\":\"$1\",\"amount\":100,\"transferId\":\"transfer-solve\"}"
}

drive_lab1() {
    temporal workflow execute "${TEMPORAL_ARGS[@]}" \
        --type Greeting --task-queue "$GREETING_TQ" --workflow-id "$GREETING_WF" \
        --input '"Temporal"'
}

drive_lab2() {
    # Approved path.
    start_transfer account-456
    wait_for_status AWAITING_APPROVAL
    temporal workflow signal "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" --name approve --input true
    timeout 60 temporal workflow result "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF"

    # Timed-out path: send nothing and let the 30 second timer fire.
    start_transfer account-456
    timeout 90 temporal workflow result "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF"
}

drive_lab3() {
    start_transfer invalid-account
    wait_for_status AWAITING_APPROVAL
    temporal workflow signal "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" --name approve --input true
    wait_for_status RETRYING
    timeout 60 temporal workflow update execute "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF" \
        --name updateToAccount --input '"account-456"'
    timeout 60 temporal workflow result "${TEMPORAL_ARGS[@]}" --workflow-id "$TRANSFER_WF"
}
