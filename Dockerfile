FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY *.go .

RUN go mod tidy
RUN go build -o puff

FROM alpine:3.22
LABEL authors="lukaszfiszer"

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/puff .

ENV GOMEMLIMIT=8MiB

ENTRYPOINT ["./puff"]