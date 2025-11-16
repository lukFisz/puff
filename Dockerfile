FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY *.go .

RUN go mod tidy
RUN go build -o cronapka

FROM alpine:latest
LABEL authors="lukaszfiszer"

RUN apk add --no-cache tzdata

WORKDIR /opt/cronapka

COPY --from=builder /app/cronapka .

ENV GOMEMLIMIT=8MiB

ENTRYPOINT ["./cronapka"]