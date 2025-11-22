FROM golang:1.25-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main/* .

RUN go build -o puff

FROM alpine:3.22
LABEL authors="lukaszfiszer"

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/puff puff

ENTRYPOINT ["./puff"]