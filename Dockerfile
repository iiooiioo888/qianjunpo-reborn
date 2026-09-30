# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git make
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /out/demo ./cmd/demo

FROM alpine:3.20
RUN apk add --no-cache ca-certificates make
WORKDIR /app
COPY --from=builder /out/demo /app/demo
COPY Makefile go.mod go.sum* ./
COPY pkg ./pkg
COPY cmd ./cmd
ENV CGO_ENABLED=0
CMD ["make", "test"]
