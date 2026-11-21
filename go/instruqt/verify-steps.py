#!/usr/bin/env python3
"""ABOUTME: Drift test for the Go track's pages. Run from anywhere:

    go/instruqt/verify-steps.py          check
    go/instruqt/verify-steps.py --fix    rewrite assignment code blocks from the snapshots

For every challenge in track/, it reads the page id from setup-workshop, then:

1. Every ```go block in assignment.md must be followed by an `apply <region>...`
   run block, and the go blocks before each apply must equal, in order, every
   occurrence of the regions it names in steps/<page>/ (imports excluded:
   apply syncs them silently). --fix rewrites those blocks to match.
2. In a scratch copy of the repo: `reset <page>` must give code that builds,
   running the page's apply commands in order must reproduce steps/<page>/
   exactly, and the result must build and vet. Page 3c also runs go test.

Needs python3 and go on PATH. Never touches your own lab folders.
"""
import os
import re
import shutil
import subprocess
import sys
import tempfile
import textwrap

HERE = os.path.dirname(os.path.abspath(__file__))
GO = os.path.dirname(HERE)
REPO = os.path.dirname(GO)
TRACK = os.path.join(HERE, "track")
STEPS = os.path.join(GO, "steps")
START = re.compile(r"^\s*(//|#)\s*region:\s*(\S+)\s*$")
END = re.compile(r"^\s*(//|#)\s*endregion\s*$")
FENCE = re.compile(r"^```(\S*)\s*$")

errors = []


def err(msg):
    errors.append(msg)
    print("FAIL:", msg)


def display(code):
    """How a region is shown in an assignment: dedented, tabs as 4 spaces."""
    return textwrap.dedent(code.expandtabs(4)).strip("\n")


def occurrences(page, name):
    root = os.path.join(STEPS, page)
    found = []
    for dirpath, _, files in sorted(os.walk(root)):
        for f in sorted(files):
            if f == "go.mod":
                continue
            lines = open(os.path.join(dirpath, f)).read().splitlines()
            inside, body = None, []
            for line in lines:
                m = START.match(line)
                if m:
                    inside, body = m.group(2), []
                elif END.match(line):
                    if inside == name:
                        found.append(display("\n".join(body)))
                    inside = None
                elif inside:
                    body.append(line)
    return found


def pages():
    for d in sorted(os.listdir(TRACK)):
        setup = os.path.join(TRACK, d, "setup-workshop")
        if not os.path.isfile(setup):
            continue
        m = re.search(r"^page_setup (\w+)", open(setup).read(), re.M)
        if not m:
            err(f"{d}/setup-workshop has no page_setup line")
            continue
        yield d, m.group(1)


def blocks(lines):
    """(lang, start, end) for each fenced block; start/end are fence line indexes."""
    out, i = [], 0
    while i < len(lines):
        m = FENCE.match(lines[i])
        if m and m.group(1):
            j = i + 1
            while j < len(lines) and lines[j].strip() != "```":
                j += 1
            out.append((m.group(1), i, j))
            i = j
        i += 1
    return out


def check_assignment(d, page, fix):
    path = os.path.join(TRACK, d, "assignment.md")
    lines = open(path).read().split("\n")
    applies, pending, edits = [], [], []
    for lang, s, e in blocks(lines):
        body = "\n".join(lines[s + 1:e])
        if lang == "go":
            pending.append((s, e, body))
            continue
        if not lang.startswith("bash") or "run" not in lang:
            continue
        for cmd in lines[s + 1:e]:
            if not cmd.startswith("apply "):
                continue
            names = cmd.split()[1:]
            applies.append(names)
            want = [o for n in names for o in occurrences(page, n)]
            for n in names:
                if not occurrences(page, n):
                    err(f"{d}: `apply {n}` but page {page} has no region {n}")
            if len(pending) != len(want):
                err(f"{d}: {len(pending)} go block(s) before `{cmd}`, page {page} has {len(want)} occurrence(s)")
            else:
                for (bs, be, got), w in zip(pending, want):
                    if got.strip("\n") != w:
                        if fix:
                            edits.append((bs, be, w))
                        else:
                            err(f"{d}: go block at line {bs + 1} does not match region in steps/{page} (run with --fix)")
            pending = []
    for bs, _, _ in pending:
        err(f"{d}: go block at line {bs + 1} is never applied")
    if edits:
        for bs, be, w in sorted(edits, reverse=True):
            lines[bs + 1:be] = w.split("\n")
        open(path, "w").write("\n".join(lines))
        print(f"fixed {len(edits)} block(s) in {d}/assignment.md")
    return applies


def run(cmd, cwd):
    p = subprocess.run(cmd, cwd=cwd, shell=True, capture_output=True, text=True)
    return p.returncode, (p.stdout + p.stderr).strip()


def main():
    fix = "--fix" in sys.argv
    plan = [(d, page, check_assignment(d, page, fix)) for d, page in pages()]

    scratch = tempfile.mkdtemp(prefix="verify-steps-")
    try:
        shutil.copytree(GO, os.path.join(scratch, "go"),
                        ignore=shutil.ignore_patterns("lab1", "lab2", "lab3", ".workshop-page"))
        shutil.copytree(os.path.join(REPO, "shared"), os.path.join(scratch, "shared"))
        g = os.path.join(scratch, "go")
        for d, page, applies in plan:
            lab = "lab" + page[0]
            rc, out = run(f"bin/reset {page}", g)
            if rc:
                err(f"{d}: reset {page} failed: {out}")
                continue
            rc, out = run(f"go build ./{lab}/... && go vet ./{lab}/...", g)
            if rc:
                err(f"{d}: start of page {page} does not build:\n{out}")
            for names in applies:
                rc, out = run("bin/apply " + " ".join(names), g)
                if rc:
                    err(f"{d}: apply {' '.join(names)} failed: {out}")
            rc, out = run(f"diff -r {lab} steps/{page}", g)
            if rc:
                err(f"{d}: the page's apply commands do not reproduce steps/{page}:\n{out}")
            rc, out = run(f"go build ./{lab}/... && go vet ./{lab}/...", g)
            if rc:
                err(f"{d}: end of page {page} does not build:\n{out}")
            if page == "3c":
                rc, out = run(f"go test -count=1 ./{lab}", g)
                if rc:
                    err(f"{d}: go test fails at the end of page 3c:\n{out}")
            if not any(e.startswith(d) for e in errors):
                print(f"ok   {d} ({page}): {len(applies)} apply command(s)")
    finally:
        shutil.rmtree(scratch)

    if errors:
        print(f"\n{len(errors)} problem(s)")
        sys.exit(1)
    print("\nAll pages verified.")


if __name__ == "__main__":
    main()
