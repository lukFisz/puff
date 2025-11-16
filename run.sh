#!/bin/zsh

mkdir -p target

go build -o target/puff

export PUFF_CRON_SCHEDULE="0/10 * * * * *"
export PUFF_DELUGE_URL="http://deluge.lan/json"
export PUFF_DELUGE_PASSWORD="deluge"
export PUFF_START_DELAY="5s"
export PUFF_DRY_RUN="true"
export PUFF_LOG_LEVEL="debug"

./target/puff