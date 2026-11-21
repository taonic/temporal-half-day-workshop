// Page snapshots. This file makes steps/ its own module so `go build ./...`
// in go/ skips it: a snapshot only compiles once `reset` or `apply` copies it
// into a lab folder. See ../../shared/steps.sh.
module github.com/taonic/temporal-half-day-workshop/go/steps

go 1.23.0
