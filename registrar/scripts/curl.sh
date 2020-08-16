#!/usr/bin/env bash
# "curl" for GRPC

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

if ! command -v grpcurl >/dev/null; then
  echo "Err: Missing grpcurl"
  exit 1
fi

grpcurl -insecure -import-path "$DIR/../api/" --proto "registrar.proto" "$@"
