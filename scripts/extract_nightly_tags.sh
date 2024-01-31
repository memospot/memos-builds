#!/usr/bin/env bash

set -eu

git config --global user.email "github-actions[bot]@users.noreply.github.com"
git config --global user.name "github-actions[bot]"

DATE_STRING=$(date +'%Y%m%d')
echo "DATE_STRING=$DATE_STRING" >>$GITHUB_ENV

PREVIOUS_TAG=$(git describe --tags --abbrev=0)
echo "PREVIOUS_TAG=$PREVIOUS_TAG" >>$GITHUB_ENV
echo "GORELEASER_PREVIOUS_TAG=$PREVIOUS_TAG" >>$GITHUB_ENV

# if [[ $PREVIOUS_TAG == *"-dev" ]]; then
#     VERSION=$(echo $PREVIOUS_TAG | awk -F. '/[0-9]+\./{$NF++;print}' OFS=.)"-dev"
#     echo "GIT_TAG=$VERSION" >>$GITHUB_ENV
#     echo "GORELEASER_CURRENT_TAG=$VERSION" >>$GITHUB_ENV
#     echo "Previous tag is a dev release. Current tag set to $VERSION"
#     exit 0
# fi

FILE_VERSION="v"$(sed -En 's/^var[[:space:]]+Version[[:space:]]+=[[:space:]]+"([[:digit:].]+)"$/\1/p' memos-upstream/server/version/version.go)

if [ -n "$FILE_VERSION" ]; then
    VERSION=$FILE_VERSION
elif [ -n "$PREVIOUS_TAG" ]; then
    VERSION=$PREVIOUS_TAG
else
    VERSION="v"$(date +%Y.%m.%d)".0"
fi
VERSION=$(echo $VERSION | awk -F. '/[0-9]+\./{$NF++;print}' OFS=.)"-dev"

echo "MEMOS_VERSION=$FILE_VERSION" >>$GITHUB_ENV
echo "GIT_TAG=$VERSION" >>$GITHUB_ENV
echo "GORELEASER_CURRENT_TAG=$VERSION" >>$GITHUB_ENV

# git tag -l "v*.*.*-dev"
# git tag -l "v*.*.*[!-dev]"
