#!/usr/bin/env bash
#
# Lint the current repository

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

golangci-lint run --fast -c "$DIR/../.golangci.yml"
