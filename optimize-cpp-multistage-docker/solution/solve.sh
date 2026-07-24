#!/bin/bash
set -euo pipefail

cd /app/src

# Fix CMakeLists.txt
echo "target_link_libraries(app mylib)" >> CMakeLists.txt

# Rewrite Dockerfile
cat << 'EOF' > Dockerfile
FROM debian:bookworm-slim@sha256:4724b8cc51e33e398f0e2e15e18d5ec2851ff0c2280647e1310bc1642182655d AS builder
RUN apt-get update \
    && apt-get install -y --no-install-recommends cmake g++ make \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /build
COPY . .
RUN cmake . && make

FROM debian:bookworm-slim@sha256:4724b8cc51e33e398f0e2e15e18d5ec2851ff0c2280647e1310bc1642182655d
RUN useradd -m appuser
USER appuser
COPY --from=builder /build/app /usr/local/bin/app
CMD ["app"]
EOF
