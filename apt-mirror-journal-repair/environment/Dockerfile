FROM golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
        python3=3.11.2-1+b1 \
        python3-pip=23.0.1+dfsg-1 \
        python3-venv=3.11.2-1+b1 \
        tmux=3.3a-3 \
        asciinema=2.2.0-1 \
        patch=2.7.6-7 \
        ca-certificates=20230311+deb12u1 \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.lock /tmp/requirements.lock

RUN python3 -m venv /opt/verifier \
    && /opt/verifier/bin/pip install --no-cache-dir -r /tmp/requirements.lock

ENV PATH=/opt/verifier/bin:${PATH}
ENV GOPATH=/tmp/go
ENV GOMODCACHE=/tmp/go/pkg/mod

RUN mkdir -p /app/bin /app/output /app/environment /logs/verifier

COPY go.mod Makefile /app/environment/
COPY archive/ /app/environment/archive/
COPY config/ /app/environment/config/
COPY data/ /app/environment/data/
COPY docs/ /app/environment/docs/
COPY engine/ /app/environment/engine/
COPY k7n/ /app/environment/k7n/
COPY m3x/ /app/environment/m3x/
COPY pilot/ /app/environment/pilot/
COPY r2p/ /app/environment/r2p/
COPY schemas/ /app/environment/schemas/
COPY tools/ /app/environment/tools/
COPY util/ /app/environment/util/

RUN cd /app/environment && go build -o /app/bin/sync_run ./pilot
