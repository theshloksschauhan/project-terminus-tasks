#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"
patch -p1 --batch -d /app/environment < patches/m3x.patch
patch -p1 --batch -d /app/environment < patches/k7n.patch
patch -p1 --batch -d /app/environment < patches/r2p.patch
patch -p1 --batch -d /app/environment < patches/engine.patch

cd /app/environment
go build -o /app/bin/sync_run ./pilot
/app/bin/sync_run
