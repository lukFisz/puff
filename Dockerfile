FROM alpine:3.22
LABEL authors="lukaszfiszer"

RUN apk add --no-cache tzdata

WORKDIR /app

ARG TARGET_ARCH="amd64"

COPY puff_${TARGET_ARCH} puff

ENTRYPOINT ["./puff"]