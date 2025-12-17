#!/bin/zsh

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
docker run -v "./target:/output" --rm puff-builder

echo "-----------------------------------"
echo "Generate preview image"
echo "-----------------------------------"
docker build \
    -f preview.Dockerfile \
    . -t ghcr.io/lukfisz/puff-preview
docker run --rm -v $PWD:/vhs ghcr.io/lukfisz/puff-preview
