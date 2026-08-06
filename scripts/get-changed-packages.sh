#!/usr/bin/env bash
#
# get-changed-packages.sh
# Returns unique Go package paths from commits ahead of main.
# Compares current branch against origin/main to find changed .go files.
# Output: one package path per line, relative to repo root (e.g. internal/service/payment)
#
# Usage:
#   ./scripts/get-changed-packages.sh              # list packages line by line
#   ./scripts/get-changed-packages.sh --join       # space-separated, for direct use with lint/test commands

set -eo pipefail

JOIN=false

if [[ "${1:-}" == "--join" ]]; then
    JOIN=true
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"

# Find the merge base between HEAD and main
# This gives us the fork point where the current branch diverged from main
merge_base=$(git merge-base main HEAD 2>/dev/null || echo "")

if [[ -z "$merge_base" ]]; then
    echo "Could not find merge base with main. Falling back to HEAD~1." >&2
    merge_base="HEAD~1"
fi

# Get changed .go files between main and HEAD (added, modified, renamed)
changed_files=$(git diff --name-only --diff-filter=ACMR "$merge_base" HEAD -- '*.go')

if [[ -z "$changed_files" ]]; then
    exit 0
fi

# Extract unique package directories from changed files
packages=$(echo "$changed_files" | xargs -I{} dirname "{}" | sort -u)

if [[ "$JOIN" == true ]]; then
    result=""
    while IFS= read -r pkg; do
        result+="./${pkg} "
    done <<< "$packages"
    echo "$result" | xargs
else
    while IFS= read -r pkg; do
        echo "./${pkg}"
    done <<< "$packages"
fi
