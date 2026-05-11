# syntax=docker/dockerfile:1.7

# CGO off so the final stage can be FROM scratch (no libc to link against).
FROM golang:1.25 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY data/ ./data/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /out/server ./cmd/server/

FROM scratch
COPY --from=build /src/data /data
COPY --from=build /out/server /server

# /state is writable at runtime (characters.json, inn.json, mail/).
# Cloudflare Containers will treat this as ephemeral; phase 4 swaps in
# an R2- or D1-backed store. Locally, bind-mount a host dir onto /state.
EXPOSE 8080
ENTRYPOINT ["/server", \
    "--port", "8080", \
    "--data-dir", "/data", \
    "--state-dir", "/state"]
