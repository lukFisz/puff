#!/bin/zsh

mkdir -p target

go build -o target/puff ./main

export PUFF_CRON_SCHEDULE="0/10 * * * * *"
export PUFF_DELUGE_URL="http://deluge.lan/json"
export PUFF_DELUGE_PASSWORD="deluge"
export PUFF_DRY_RUN="true"
export PUFF_RETENTION="P5D"
export PUFF_LOG_LEVEL="debug"
export PUFF_DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/1392149317812486358/hDQ2v2kzTqITcr_Qt5oQEsaFBu7woFq5lu2jDn_fFfTSfBVtKqDYtGDbHHY2SxITbyTM"

./target/puff