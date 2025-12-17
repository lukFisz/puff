FROM golang:1.25-alpine3.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal/ internal/
COPY main.go .

RUN go test ./...

RUN GOOS="linux" GOARCH="amd64" go build -o puff-linux-amd64
RUN GOOS="linux" GOARCH="arm64" go build -o puff-linux-arm64

CMD ["cp", "puff-linux-arm64", "puff-linux-amd64", "/output"]
