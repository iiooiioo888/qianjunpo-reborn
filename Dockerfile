# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git make
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /out/demo ./cmd/demo

FROM golang:1.22-alpine
RUN apk add --no-cache ca-certificates git make
WORKDIR /app
COPY --from=builder /out/demo /app/demo
COPY Makefile go.mod go.sum* ./
COPY pkg ./pkg
COPY internal ./internal
COPY gen ./gen
COPY services ./services
COPY cmd ./cmd
ENV CGO_ENABLED=0
CMD ["make", "test"]
