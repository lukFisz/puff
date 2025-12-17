#!/bin/bash

set -e

export PVER=$1

if [ -z "$PVER" ]; then
    echo "Specify version:"
    echo "> ./preview.sh VERSION"
    exit 1
fi

echo "-----------------------------------"
echo "Puff builder"
echo "-----------------------------------"
docker build \
    -f builder.Dockerfile \
    . -t ghcr.io/lukfisz/puff-builder

echo "-----------------------------------"
echo "Generate preview image"
echo "-----------------------------------"
docker build \
    --build-arg PUFF_CURRENT_VERSION=$PVER \
    -f preview.Dockerfile \
    . -t ghcr.io/lukfisz/puff-preview
docker run --rm \
    -v $PWD:/mnt/puff/monitor \
    -v $PWD:/vhs \
    ghcr.io/lukfisz/puff-preview
