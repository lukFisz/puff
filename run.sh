#!/bin/zsh

set -e

mkdir -p target

go build -o target/puff

export PUFF_CRON_SCHEDULE="0/10 * * * * *"
export PUFF_DELUGE_URL="http://deluge.lan/json"
export PUFF_DELUGE_PASSWORD="deluge"
export PUFF_DRY_RUN="true"
export PUFF_RETENTION="P5D"
export PUFF_LOG_LEVEL="debug"
export PUFF_RUN_ONCE="false"
export TZ="Europe/Warsaw"

./target/puff
