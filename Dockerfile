# syntax=docker/dockerfile:1.7

# ---- build stage ----------------------------------------------------------
# Compiles a fully static Linux binary. CGO is off so we don't pull in libc,
# which lets the final stage be FROM scratch.
FROM golang:1.25 AS build

WORKDIR /src

# Cache module downloads in their own layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -s -w strips debug/symbol tables (smaller binary). GOOS=linux is explicit
# in case someone builds on a non-Linux Docker host.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /out/server ./cmd/server/

# ---- final stage ----------------------------------------------------------
# `scratch` is a literal empty image — no shell, no package manager, no libc.
# The container is just our binary plus the read-only data/ directory.
FROM scratch

# Read-only game assets (monsters, weapons, terrain, ANSI screens).
COPY --from=build /src/data /data

# Static server binary.
COPY --from=build /out/server /server

# /state is the writable directory for characters.json, inn.json, mail/.
# In Cloudflare Containers this will be ephemeral; phase 4 swaps in R2/D1.
# Locally you can bind-mount a host directory onto /state for persistence.

EXPOSE 8080

ENTRYPOINT ["/server", \
    "--port", "8080", \
    "--data-dir", "/data", \
    "--state-dir", "/state"]
