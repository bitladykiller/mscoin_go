# syntax=docker/dockerfile:1.7

# UCenter API owns its own Dockerfile so Compose can build it directly.
# This keeps the HTTP adapter isolated from the rest of the stack while still
# allowing it to consume the shared module and generated RPC contracts.
ARG GO_IMAGE=golang:1.26.3
FROM ${GO_IMAGE} AS builder

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /workspace

COPY go.mod go.sum go.work go.work.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY app ./app
COPY idl ./idl
COPY pkg ./pkg

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/server ./app/ucenter/api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY --from=builder /workspace/app/ucenter/api/etc /app/etc

ENTRYPOINT ["/app/server"]
CMD ["-f", "/app/etc/ucenter-api.yaml"]
