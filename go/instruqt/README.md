# Instruqt: Temporal Half-Day Workshop (Go)

The three Go labs packaged as an Instruqt track. Attendees get a container
with a Temporal dev server running, Go 1.24 with the SDK already in the build
cache, the editor open on the lab package, and the solution one tab away.

```
go/instruqt/
├── track/        3 challenge directories + track.yml   ->  instruqt track push
├── sandbox/      config.yml + scripts/                 ->  instruqt sandbox push
└── lib-go.sh     Go helpers for the lifecycle scripts  ->  git push
../../shared/instruqt/lib.sh   checks and CLI drivers   ->  git push
```

| Track | `temporal-half-day-workshop-go` |
|---|---|
| Preset | `temporal-half-day-sandbox-go` |
| Repo cloned by the preset | `https://github.com/taonic/temporal-half-day-workshop.git` at `WORKSHOP_REF` (default `main`) |

## Three ways a change ships

| What you edited | How it ships | Live at |
|---|---|---|
| `track/`: assignments, lifecycle scripts, `track.yml` | `instruqt track push` | next lab start |
| `sandbox/config.yml`, `sandbox/scripts/` | `instruqt sandbox push` then `publish` | next sandbox provision |
| `lab*/`, `solution*/`, `lib-go.sh`, `shared/instruqt/` | `git push` to `WORKSHOP_REF` | next sandbox provision |

> [!IMPORTANT]
> The third row catches people. Lab code and both helper libraries reach the
> sandbox only through the `git clone` in `sandbox/scripts/setup-workshop`.
> Pushing the track does not update them.

The lifecycle scripts are thin: each one sources `shared/instruqt/lib.sh`
(language-neutral checks and CLI drivers, see [`SPEC.md`](../../SPEC.md)) and
`lib-go.sh` (killing Go processes, the build gate, copying a solution in).
The helpers live outside `track/` because Instruqt parses every file inside a
challenge directory as a lifecycle script.

## What each check requires

Every check first requires the lab's TODOs to be gone and the package to
compile. Runs only count if they started after the challenge's setup, which
writes `/tmp/workshop-start-NN`.

| Lab | Then it requires |
|---|---|
| 1 | A `RetryPolicy` in `workflow.go`; `greeting-workflow` returned `Hello, Temporal!` with `ActivityTaskStarted.attempt == 3` |
| 2 | Query handler, Signal channel and `AwaitWithTimeout` in `workflow.go`; one run `COMPLETED` with an `approve` Signal, one `TIMED_OUT` with `TimerFired` and no Signal |
| 3 | Non-retryable error, Update handler and `AllHandlersFinished`; tests assert `StatusTimedOut` and send the Update; `go test ./lab3` passes 2+ tests; one run `COMPLETED` with a non-retryable `ActivityTaskFailed` and an accepted and completed `updateToAccount` |

Solve scripts copy `solutionN/` into `labN/` with the package name rewritten,
start the Worker, and play the learner's part through the CLI. Lab 2's solve
takes about 35 seconds because it waits out the real 30-second timer.

## Test locally before pushing

`instruqt track validate` never runs the scripts. Run them in a container
seeded with your working tree, which makes `setup-workshop` skip the clone.
Run from the repo root. Instruqt runs `linux/amd64`; add
`--platform linux/amd64` to `docker run` to match it exactly, or leave it off
for a faster native run on Apple silicon (the script handles both).

```bash
docker rm -f hdw-test 2>/dev/null
docker run -d --name hdw-test -p 8299:8233 ubuntu:24.04 sleep infinity
docker exec hdw-test mkdir -p /opt/instruqt/bootstrap /root/workshop
docker exec hdw-test touch /opt/instruqt/bootstrap/host-bootstrap-completed
tar --exclude='./.git' -cf - . | docker exec -i hdw-test tar -xf - -C /root/workshop
docker cp go/instruqt/sandbox/scripts/setup-workshop hdw-test:/tmp/setup-workshop
docker exec hdw-test /tmp/setup-workshop

docker cp go/instruqt/track hdw-test:/tmp/ch
docker exec hdw-test env -u HOME bash -c '
  for d in /tmp/ch/0*; do
    echo "## $(basename $d)"
    $d/setup-workshop
    $d/check-workshop && echo "!! check passed before solve"
    $d/solve-workshop >/dev/null 2>&1
    $d/check-workshop || echo "!! check failed after solve"
    $d/cleanup-workshop
  done'

docker rm -f hdw-test
```

The Temporal UI is on <http://localhost:8299> while the container runs.

## First publish

Neither the preset nor the track exists on Instruqt yet. The preset must
exist before the track validates, and the repo must be public and pushed
before the preset can provision.

```bash
git push                                   # lab code and helpers on WORKSHOP_REF
cd go/instruqt/sandbox
instruqt sandbox push                      # creates the preset as a draft
instruqt sandbox publish --message="Initial Go half-day sandbox"
cd ../track
instruqt track validate
instruqt track push                        # creates the track, writes ids into the YAML
git add . && git commit -m "Track: record Instruqt ids"
instruqt track test
```

`track.yml` and the assignments ship without `id:` and `checksum`. The first
push writes them back; commit that. Every later push rewrites the frontmatter
(comments stripped, keys reordered), so commit it as-is.

> [!WARNING]
> `instruqt sandbox publish` moves every track on the preset to the new
> version. This preset serves only this track, but still: never publish
> during a live session.

## Live workshop pre-flight

- [ ] `WORKSHOP_REF` pinned to a tag or commit in `sandbox/scripts/setup-workshop`, preset re-published.
- [ ] Lab and helper changes pushed to that ref.
- [ ] `maintenance: false` in `track.yml`, pushed. It ships `true`, so only owners and authors can launch.
- [ ] Hot Start pool sized to attendees plus 10 to 20 percent, warmed after the last change. Launch one pooled sandbox and click through a lab.
- [ ] `idle_timeout` set in the web UI (Track, Settings) to workshop length plus 20 minutes. The web UI wins over the pushed value.
- [ ] `feedback_recap_enabled` / `feedback_tab_enabled` set to `false` if feedback is collected elsewhere.
- [ ] `instruqt track test` passes.

Afterwards, disable Hot Start. Idle pooled sandboxes cost money.
