#!/usr/bin/env bash
# Tag the current HEAD and push so the GitHub release workflows fire.
# Auto-bumps the patch of the latest v<major>.<minor>.<patch> tag unless
# a tag is supplied: ./release.sh v1.0.0
set -euo pipefail

cd "$(dirname "$0")"

if [[ -n "$(git status --porcelain)" ]]; then
    echo "error: working tree has changes; commit or stash before tagging" >&2
    git status --short >&2
    exit 1
fi

BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "dev" ]]; then
    echo "error: on branch '$BRANCH'; release from dev" >&2
    exit 1
fi

git fetch --quiet --tags origin dev
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/dev)
if [[ "$LOCAL" != "$REMOTE" ]]; then
    echo "error: local dev ($LOCAL) is not in sync with origin/dev ($REMOTE)" >&2
    exit 1
fi

if [[ $# -ge 1 ]]; then
    TAG="$1"
else
    LAST=$(git tag --list 'v*' --sort=-v:refname | head -n1)
    if [[ -z "$LAST" ]]; then
        TAG="v0.0.1"
    else
        IFS=. read -r MAJOR MINOR PATCH <<<"${LAST#v}"
        TAG="v${MAJOR}.${MINOR}.$((PATCH + 1))"
    fi
fi

if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
    echo "error: tag ${TAG} already exists" >&2
    exit 1
fi

echo "Tagging $(git rev-parse --short HEAD) as ${TAG}"
git tag -a "${TAG}" -m "Release ${TAG}"
git push origin "${TAG}"
echo "Pushed ${TAG}. Release workflows should now run:"
echo "  https://github.com/fisherevans/project-f/actions"
