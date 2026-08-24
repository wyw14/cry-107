#!/usr/bin/env sh
set -eu
docker build -f benzhi.Dockerfile -t kilnguard:local .
