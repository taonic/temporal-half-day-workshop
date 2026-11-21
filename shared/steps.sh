#!/bin/bash
# ABOUTME: `apply` and `reset` for the workshop pages. Language-neutral.
#
#   steps.sh <lang-dir> apply [page] <region>...
#   steps.sh <lang-dir> reset [page]
#   steps.sh <lang-dir> restore <snapshot>     solve scripts and cache warm-up
#   steps.sh <lang-dir> page                   print the current page
#
# Each language folder has steps/<snapshot>/, a full copy of one lab folder:
# steps/1-start is Lab 1 before page 1a, steps/1a is Lab 1 after page 1a, and
# so on. Learners work in <lang-dir>/lab1, lab2, lab3.
#
# Code that a page changes sits between markers, in any language that has
# `//` or `#` comments:
#
#   // region: retry-policy
#   ...
#   // endregion
#
# `apply retry-policy` copies every region of that name from the current
# page's snapshot into the lab folder, plus the file's `imports` region, and
# leaves the rest of the file alone. Regions do not nest. A name may occur in
# several places and files; the occurrences pair up in order.
#
# `reset 1b` puts the lab folder back to the start of page 1b and makes 1b
# the current page. Lifecycle setup scripts call it; learners can too.
set -euo pipefail

ROOT=$(cd "${1:?usage: steps.sh <lang-dir> <command> ...}" && pwd)
shift
CMD=${1:-help}
shift || true
STEPS=$ROOT/steps
STATE=$ROOT/.workshop-page

die() { echo "$*" >&2; exit 1; }
is_page() { [[ ${1:-} =~ ^[0-9][a-z]$ ]]; }
lab_of() { echo "lab${1:0:1}"; }

# prev_of 1a -> 1-start, prev_of 1c -> 1b
prev_of() {
    local n=${1:0:1} l=${1:1:1}
    if [ "$l" = a ]; then echo "$n-start"; else echo "$n$(echo "$l" | tr 'b-z' 'a-y')"; fi
}

current_page() {
    local page
    page=$(cat "$STATE" 2>/dev/null || true)
    is_page "$page" || die "No page selected yet. Start one with: reset 1a"
    echo "$page"
}

restore() {
    local snap=$1 lab
    [ -d "$STEPS/$snap" ] || die "No snapshot $snap in $STEPS."
    lab=$(lab_of "$snap")
    rm -rf "${ROOT:?}/$lab"
    cp -R "$STEPS/$snap" "$ROOT/$lab"
}

# Marker lines, for grep and awk alike.
START_RE='^[[:space:]]*(//|#)[[:space:]]*region:[[:space:]]*'
END_RE='^[[:space:]]*(//|#)[[:space:]]*endregion[[:space:]]*$'

# splice <source> <target> <name>...: print target with each named region's
# body taken from source. Exit 3 if the markers in the two files do not
# line up: a missing, extra, unclosed or nested marker.
splice() {
    awk -v names=" ${*:3} " -v sre="$START_RE" -v ere="$END_RE" '
        function region_name(line) { sub(sre, "", line); sub(/[[:space:]]+$/, "", line); return line }
        function wanted(n) { return index(names, " " n " ") > 0 }
        FNR == 1 { file++ ; inside = "" }
        file == 1 {
            if ($0 ~ sre) {
                if (inside != "") exit 3
                inside = region_name($0)
                if (wanted(inside)) { k = inside SUBSEP (++have[inside]); body[k] = "" }
                next
            }
            if ($0 ~ ere) { if (inside == "") exit 3; inside = ""; next }
            if (inside != "" && wanted(inside)) body[k] = body[k] $0 "\n"
            next
        }
        {
            if ($0 ~ sre) {
                if (inside != "") exit 3
                inside = region_name($0)
                print
                if (wanted(inside)) {
                    if (++seen[inside] > have[inside]) exit 3
                    printf "%s", body[inside SUBSEP seen[inside]]
                    skipping = 1
                }
                next
            }
            if ($0 ~ ere) {
                if (inside == "") exit 3
                inside = ""; skipping = 0
                print
                next
            }
            if (!skipping) print
        }
        END {
            if (inside != "") exit 3
            for (n in have) if (seen[n] != have[n]) exit 3
        }
    ' "$1" "$2"
}

cmd_apply() {
    local page names=() n src lab files f rel tgt here worker=false
    if is_page "${1:-}"; then page=$1; shift; else page=$(current_page); fi
    [ $# -gt 0 ] || die "usage: apply [page] <region>..."
    names=("$@")
    src=$STEPS/$page
    lab=$(lab_of "$page")
    [ -d "$src" ] || die "No page $page in $STEPS."
    [ -d "$ROOT/$lab" ] || die "$lab is missing. Start the page with: reset $page"

    # Every file on this page that has one of the named regions.
    files=$(cd "$src" && for n in "${names[@]}"; do
        grep -rlE "${START_RE}${n}[[:space:]]*$" . || true
    done | sort -u)
    for n in "${names[@]}"; do
        (cd "$src" && grep -rqE "${START_RE}${n}[[:space:]]*$" .) || {
            echo "Page $page has no region named \"$n\". Its regions:" >&2
            (cd "$src" && grep -rhE "$START_RE" . | sed -E "s@${START_RE}@  @" | grep -vx '  imports' | sort -u) >&2
            exit 1
        }
    done

    # Build every new file first, so a bad marker changes nothing at all.
    APPLY_TMP=$(mktemp -d)
    trap 'rm -rf "$APPLY_TMP"' EXIT
    local tmp=$APPLY_TMP
    for f in $files; do
        rel=${f#./}
        tgt=$ROOT/$lab/$rel
        [ -f "$tgt" ] || die "$lab/$rel is missing. Put the page back with: reset $page"
        mkdir -p "$tmp/$(dirname "$rel")"
        if ! splice "$src/$rel" "$tgt" "${names[@]}" imports > "$tmp/$rel"; then
            echo "The region markers in $lab/$rel no longer match page $page." >&2
            echo "Nothing was changed. Put the page back with: reset $page" >&2
            echo "then run: apply ${names[*]}" >&2
            exit 1
        fi
    done
    for f in $files; do
        rel=${f#./}
        if cmp -s "$tmp/$rel" "$ROOT/$lab/$rel"; then
            echo "$lab/$rel: already up to date"
            continue
        fi
        cp "$tmp/$rel" "$ROOT/$lab/$rel"
        here=$(for n in "${names[@]}"; do grep -qE "${START_RE}${n}[[:space:]]*$" "$src/$rel" && printf '%s ' "$n"; done; true)
        echo "$lab/$rel: applied ${here% }"
        case $rel in *_test.* | *test_*) ;; *) worker=true ;; esac
    done
    if $worker; then
        echo "Restart the Worker to run the new code: Ctrl+C in the Worker tab, then press Up and Enter."
    fi
}

cmd_reset() {
    local page
    if is_page "${1:-}"; then page=$1; else page=$(current_page); fi
    [ -d "$STEPS/$page" ] || die "No page $page in $STEPS."
    restore "$(prev_of "$page")"
    echo "$page" > "$STATE"
    echo "$(lab_of "$page") is at the start of page $page."
}

case $CMD in
    apply) cmd_apply "$@" ;;
    reset) cmd_reset "$@" ;;
    restore) restore "${1:?usage: restore <snapshot>}" ;;
    page) current_page ;;
    *) die "usage: steps.sh <lang-dir> apply [page] <region>... | reset [page] | restore <snapshot> | page" ;;
esac
