# Instruqt: Temporal Half-Day Workshop (Go)

The Go labs packaged as an Instruqt track: nine pages, one challenge each.
Attendees get a container with a Temporal dev server running, Go 1.24 with
the SDK already in the build cache, and VS Code (code-server) open on the lab
folder. Every page shows its code, and a run button on `apply <region>`
writes it into the editor.

```
go/
├── steps/          a snapshot of the lab folder before and after every page
├── bin/            apply, reset: wrappers around ../shared/steps.sh
└── instruqt/
    ├── track/          9 challenge directories + track.yml   ->  instruqt track push
    ├── sandbox/        config.yml + scripts/                 ->  instruqt sandbox push
    ├── lib-go.sh       Go helpers for the lifecycle scripts  ->  git push
    └── verify-steps.py drift test for assignments and snapshots
../shared/steps.sh              apply and reset, language-neutral  ->  git push
../shared/instruqt/lib.sh       checks and CLI drivers             ->  git push
```

| Track | `temporal-half-day-workshop-go` |
|---|---|
| Preset | `temporal-half-day-sandbox-go` |
| Repo cloned by the preset | `https://github.com/taonic/temporal-half-day-workshop.git` at `WORKSHOP_REF` (default `main`) |

## How a page works

1. The challenge's setup runs `reset <page>`: it stops every Worker,
   terminates leftover Workflows, and recreates `labN/` from the snapshot
   before the page. Everyone starts from code that builds, whatever they did
   on the last page.
2. The assignment shows each region's code, then a run block `apply <region>`.
   `apply` copies those regions (and the file's `imports` region) from
   `steps/<page>/` into `labN/`. code-server reloads the open file.
3. The learner restarts the Worker, runs the page, and tries the experiment.
4. Check looks for the page's run in the Temporal history. Nothing else.

Regions, snapshots and the commands are specified in [SPEC.md](../../SPEC.md).

## Three ways a change ships

| What you edited | How it ships | Live at |
|---|---|---|
| `track/`: assignments, lifecycle scripts, `track.yml` | `instruqt track push` | next lab start |
| `sandbox/config.yml`, `sandbox/scripts/` | `instruqt sandbox push` then `publish` | next sandbox provision |
| `steps/`, `bin/`, `lib-go.sh`, `shared/` | `git push` to `WORKSHOP_REF` | next sandbox provision |

> [!IMPORTANT]
> The third row catches people. Snapshots and helpers reach the sandbox only
> through the `git clone` in `sandbox/scripts/setup-workshop`. Pushing the
> track does not update them.

Every lifecycle script is one line, `page_setup 1b` and so on, after sourcing
`shared/instruqt/lib.sh` and `lib-go.sh`. The helpers live outside `track/`
because Instruqt parses every file inside a challenge directory as a
lifecycle script.

## Changing a page's code

1. Edit the snapshot in `steps/<page>/`. A change to a region that later
   pages also carry goes into those snapshots too.
2. Run `go/instruqt/verify-steps.py --fix`. It rewrites the assignments'
   `go` blocks from the snapshots, then checks that every page builds at its
   start and end, and that its `apply` commands reproduce its snapshot.
3. Commit and `git push`.

## What each check requires

Runs only count if they started after the page's setup, which writes
`/tmp/workshop-start-<page>`.

| Page | Passes when |
|---|---|
| 1a | `greeting-workflow` returned `Hello, Temporal!` |
| 1b | `greeting-workflow` failed after `ActivityTaskStarted.attempt == 2` |
| 1c | `Hello, Temporal!` with `ActivityTaskStarted.attempt == 3` |
| 2a | A transfer `COMPLETED` with an `approve` Signal |
| 2b | As 2a, and the learner's running Worker answers `getStatus` with `COMPLETED` |
| 2c | A transfer `TIMED_OUT` with `TimerFired` and no Signal |
| 3a | A transfer failed with a non-retryable `ActivityTaskFailed` |
| 3b | A transfer `COMPLETED` with a non-retryable failure and an accepted, completed `updateToAccount` |
| 3c | `go test ./lab3` passes both tests, none skipped |

Solve scripts restore the page's finished snapshot, start its Worker (left
running for page 2b's check; the next setup stops it), and play the
learner's part through the CLI. Page 2c's solve takes about 35 seconds
because it waits out the real 30-second timer.

## Test locally before pushing

`instruqt track validate` never runs the scripts. First the drift test, on
your machine (needs `python3` and `go`):

```bash
go/instruqt/verify-steps.py
```

Then the lifecycle scripts, in a container seeded with your working tree,
which makes `setup-workshop` skip the clone. Run from the repo root.
Instruqt runs `linux/amd64`; add `--platform linux/amd64` to `docker run`
to match it exactly, or leave it off for a faster native run on Apple
silicon (the script handles both).

```bash
docker rm -f hdw-test 2>/dev/null
docker run -d --name hdw-test -p 8299:8233 -p 8297:8080 ubuntu:24.04 sleep infinity
docker exec hdw-test mkdir -p /opt/instruqt/bootstrap /root/workshop
docker exec hdw-test touch /opt/instruqt/bootstrap/host-bootstrap-completed
tar --exclude='./.git' --exclude='./go/lab*' -cf - . | docker exec -i hdw-test tar -xf - -C /root/workshop
docker cp go/instruqt/sandbox/scripts/setup-workshop hdw-test:/tmp/setup-workshop
docker exec hdw-test /tmp/setup-workshop

docker cp go/instruqt/track hdw-test:/tmp/ch
cat > /tmp/lifecycle.sh <<'EOF'
for d in /tmp/ch/0*; do
  echo "## $(basename $d)"
  $d/setup-workshop
  $d/check-workshop && echo "!! check passed before solve"
  $d/solve-workshop >/dev/null 2>&1
  $d/check-workshop || echo "!! check failed after solve"
  $d/cleanup-workshop
done
EOF
docker cp /tmp/lifecycle.sh hdw-test:/tmp/lifecycle.sh
docker exec hdw-test env -u HOME bash /tmp/lifecycle.sh

docker rm -f hdw-test
```

Run the loop from a script file, not `bash -c '...'`: `kill_go_processes`
matches `go run` anywhere in a command line, including yours.

The Temporal UI is on <http://localhost:8299> and code-server on
<http://localhost:8297> while the container runs.

## First publish

The preset must exist before the track validates, and the repo must be
public and pushed before the preset can provision.

```bash
git push                                   # snapshots and helpers on WORKSHOP_REF
cd go/instruqt/sandbox
instruqt sandbox push                      # creates the preset as a draft
instruqt sandbox publish --message="Initial Go half-day sandbox"
cd ../track
instruqt track validate
instruqt track push                        # writes challenge and tab ids into the YAML
git add . && git commit -m "Track: record Instruqt ids"
instruqt track test
```

Challenges and tabs ship without `id:`. A push writes them back; commit
that. Every later push rewrites the frontmatter (comments stripped, keys
reordered), so commit it as-is.

> [!WARNING]
> `instruqt sandbox publish` moves every track on the preset to the new
> version. This preset serves only this track, but still: never publish
> during a live session.

## Live workshop pre-flight

- [ ] `WORKSHOP_REF` pinned to a tag or commit in `sandbox/scripts/setup-workshop`, preset re-published.
- [ ] Snapshot and helper changes pushed to that ref.
- [ ] `maintenance: false` in `track.yml`, pushed. It ships `true`, so only owners and authors can launch.
- [ ] Hot Start pool sized to attendees plus 10 to 20 percent, warmed after the last change. Launch one pooled sandbox and click through a page: the run button must land `apply` in the Terminal tab, and code-server must show the change.
- [ ] `idle_timeout` set in the web UI (Track, Settings) to workshop length plus 20 minutes. The web UI wins over the pushed value.
- [ ] `feedback_recap_enabled` / `feedback_tab_enabled` set to `false` if feedback is collected elsewhere.
- [ ] `instruqt track test` passes.

Afterwards, disable Hot Start. Idle pooled sandboxes cost money.
