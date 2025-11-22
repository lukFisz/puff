FROM alpine:3.22
LABEL authors="lukaszfiszer"

RUN apk add --no-cache tzdata

WORKDIR /app

ARG TARGETARCH

COPY puff_${TARGETARCH} puff_${TARGETARCH}

ENTRYPOINT ["./puff_${TARGETARCH}"]