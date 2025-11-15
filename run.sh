#!/bin/zsh

mkdir -p target

go build -o target/cronapka

BRUSH_CRON_SCHEDULE="0/10 * * * * *" ./target/cronapka