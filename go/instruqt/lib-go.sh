#!/bin/bash
# ABOUTME: Go-specific helpers for this track's lifecycle scripts.
#
# Sourced at runtime from the sandbox's clone, after the shared library:
#
#   . /root/workshop/shared/instruqt/lib.sh
#   . /root/workshop/go/instruqt/lib-go.sh
#
# Every challenge's lifecycle scripts are one line each: page_setup,
# page_check, page_solve or page_cleanup with the page id.
#
# It lives outside track/ on purpose: Instruqt parses every file inside a
# challenge directory as a lifecycle script.

export HOME="${HOME:-/root}" GOTOOLCHAIN=local
export PATH="/usr/local/go/bin:/usr/local/bin:$PATH"

GO_DIR=${GO_DIR:-/root/workshop/go}
STEPS_SH="$GO_DIR/../shared/steps.sh"

# `go run` execs a binary it built, and killing the go command does not kill
# that child. Go 1.24+ caches the binary at ~/.cache/go-build/<hash>-d/<name>;
# older Go used /tmp/go-build*/b001/exe/<name>. Match both. Solve scripts run
# prebuilt binaries from /tmp/solve/.
#
# An orphaned Lab 2 Worker shares money-transfer-task-queue with Lab 3 and
# silently runs Lab 3's Workflows with Lab 2's Deposit, which never fails.
# Wait for every process to exit; SIGKILL stragglers.
GO_PROC_PATTERN='go run|go-build.*/(worker|starter)$|/tmp/solve/'
kill_go_processes() {
    pkill -f "$GO_PROC_PATTERN" 2>/dev/null || true
    for _ in $(seq 1 10); do
        pgrep -f "$GO_PROC_PATTERN" >/dev/null 2>&1 || return 0
        sleep 0.5
    done
    pkill -9 -f "$GO_PROC_PATTERN" 2>/dev/null || true
    sleep 0.5
}

# start_solve_worker <lab>: build the lab's Worker and run it in the
# background. It is left running so the check can query it (page 2b); the
# next page's setup stops it.
start_solve_worker() {
    mkdir -p /tmp/solve
    go build -o "/tmp/solve/worker-$1" "./$1/worker" || return 1
    nohup "/tmp/solve/worker-$1" > "/tmp/solve-$1-worker.log" 2>&1 &
    sleep 3
}

lab_of() { echo "lab${1:0:1}"; }

# page_setup <page>: stop stragglers, put the lab folder at the start of the
# page, and stamp the start time so the check counts only this page's runs.
page_setup() {
    kill_go_processes
    terminate_workshop_workflows "Page $1 setup"
    bash "$STEPS_SH" "$GO_DIR" reset "$1" || return 1
    stamp_start "$1"
}

# page_check <page>
page_check() {
    cd "$GO_DIR" || fail "Workshop directory is missing."
    case $1 in
        3c) check_tests ;;
        *) "check_$1" "$(since "$1")" ;;
    esac
}

check_tests() {
    local out passed
    if ! out=$(go test -count=1 -v ./lab3 2>&1); then
        echo "$out" | grep -E '^(---|\s+workflow_test|FAIL|panic)|\.go:[0-9]+' | tail -20
        fail "go test ./lab3 failed. Read the failure above, apply test-timeout test-update again if needed, then check again."
    fi
    passed=$(grep -c '^--- PASS' <<<"$out")
    if [ "${passed:-0}" -lt 2 ]; then
        fail "Only ${passed:-0} of 2 tests ran. Apply test-timeout test-update, then run go test -v ./lab3."
    fi
    pass "TestApprovalTimesOut and TestInvalidAccountFixedByUpdate pass"
}

# page_solve <page>: finish the page's code, start its Worker, and play the
# learner's part from the CLI.
page_solve() {
    local lab
    lab=$(lab_of "$1")
    cd "$GO_DIR" || return 0
    kill_go_processes
    terminate_workshop_workflows "Page $1 solve"
    bash "$STEPS_SH" "$GO_DIR" restore "$1"
    [ "$1" = 3c ] && return 0
    start_solve_worker "$lab" || return 0
    "drive_$1"
}

page_cleanup() {
    kill_go_processes
}
