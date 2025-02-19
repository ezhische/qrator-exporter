FROM golang:1.23 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY cmd/*.go cmd/
COPY internal internal

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -ldflags '-s -w' -o qrator-exporter cmd/*.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/qrator-exporter .
USER 65532:65532

ENTRYPOINT ["/qrator-exporter"]
