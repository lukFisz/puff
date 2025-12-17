FROM ghcr.io/lukfisz/puff-builder AS builder

FROM ghcr.io/charmbracelet/vhs:v0.10.0

COPY --from=builder /app/puff-linux-amd64 /usr/bin/puff

COPY preview.tape .

ARG PUFF_CURRENT_VERSION
ENV PUFF_CURRENT_VERSION=$PUFF_CURRENT_VERSION
ENV PUFF_PREVIEW_MODE="true"
ENV PUFF_DISK_FREE_SPACE_THRESHOLD="1PB"
ENV PUFF_CRON_SCHEDULE="@every 5s"
ENV PUFF_DELUGE_URL="http://preview.example.com"
ENV PUFF_DELUGE_PASSWORD="preview"
ENV PUFF_DRY_RUN="true"
ENV PUFF_RETENTION="P10D"
ENV PUFF_LOG_LEVEL="debug"
ENV PUFF_RUN_ONCE="false"
ENV TZ="Europe/Warsaw"

CMD [ "preview.tape" ]
