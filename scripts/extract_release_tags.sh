#!/usr/bin/env bash

set -eu
git config --global user.email "github-actions[bot]@users.noreply.github.com"
git config --global user.name "github-actions[bot]"
echo "DATE_STRING=$(date +'%Y%m%d')" >>$GITHUB_ENV

FILE_VERSION="v"$(sed -En 's/^var[[:space:]]+Version[[:space:]]+=[[:space:]]+"([[:digit:].]+)"$/\1/p' memos-upstream/server/version/version.go)

GIT_TAG=$(git describe --tags --abbrev=0)
if [ -n "$FILE_VERSION" ]; then
    VERSION=$FILE_VERSION
elif [ -n "$GIT_TAG" ]; then
    VERSION=$GIT_TAG
else
    VERSION=${GITHUB_REF_NAME#release/}
fi

if [ -z "$VERSION" ]; then
    VERSION="v"$(date +%Y.%m.%d)".0"
fi

echo "VERSION=$VERSION" >>$GITHUB_ENV
echo "GIT_TAG=$GIT_TAG" >>$GITHUB_ENV
echo "GORELEASER_CURRENT_TAG=$VERSION" >>$GITHUB_ENV
