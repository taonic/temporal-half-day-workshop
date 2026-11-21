#!/bin/bash
# ABOUTME: Go-specific helpers for this track's lifecycle scripts.
#
# Sourced at runtime from the sandbox's clone, after the shared library:
#
#   . /root/workshop/shared/instruqt/lib.sh
#   . /root/workshop/go/instruqt/lib-go.sh
#
# It lives outside track/ on purpose: Instruqt parses every file inside a
# challenge directory as a lifecycle script.

export HOME="${HOME:-/root}" GOTOOLCHAIN=local
export PATH="/usr/local/go/bin:/usr/local/bin:$PATH"

GO_DIR=/root/workshop/go

# `go run` compiles to /tmp/go-build*/exe/<name> and execs it, so kill both
# the go command and the binary it spawned. Solve scripts run prebuilt
# binaries from /tmp/solve/. One pattern per line.
kill_go_processes() {
    pkill -f "go run"       2>/dev/null || true
    pkill -f "/exe/worker"  2>/dev/null || true
    pkill -f "/exe/starter" 2>/dev/null || true
    pkill -f "/tmp/solve/"  2>/dev/null || true
    sleep 1
}

# go_build_gate <lab>: fail the check with the compiler output if it breaks.
go_build_gate() {
    local out
    if ! out=$(go build "./$1/..." 2>&1); then
        echo "$out" | tail -15
        fail "$1 does not compile. Fix the errors above, then check again."
    fi
    pass "$1 compiles"
}

# require_code <file> <fixed-string> <message>
require_code() {
    grep -qF "$2" "$1" 2>/dev/null || fail "$3"
}

# copy_solution <n>: solutionN -> labN, renaming the package and import path.
copy_solution() {
    local n=$1 f
    (cd "solution$n" && find . -name '*.go') | while read -r f; do
        mkdir -p "lab$n/$(dirname "$f")"
        sed "s/solution$n/lab$n/g" "solution$n/$f" > "lab$n/$f"
    done
}

# start_solve_worker <n>: build labN's Worker and run it in the background.
# Built up front so the PID is the Worker itself, not a `go run` wrapper.
start_solve_worker() {
    mkdir -p /tmp/solve
    go build -o "/tmp/solve/worker-$1" "./lab$1/worker" || return 1
    "/tmp/solve/worker-$1" > "/tmp/solve-$1-worker.log" 2>&1 &
    SOLVE_WORKER_PID=$!
    sleep 3
}
