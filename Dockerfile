FROM alpine:3.22
LABEL authors="lukaszfiszer"

RUN apk add --no-cache tzdata

WORKDIR /app

COPY puff_${TARGETARCH} puff

ENTRYPOINT ["./puff"]