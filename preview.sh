#!/bin/zsh

set -e

export PVER=$1

if [ -z "$PVER" ]; then
    echo "Specify version:"
    echo "> ./preview.sh VERSION"
    exit 1
fi

echo "-----------------"
echo "Builder"
echo "-----------------"
docker build \
    -f Dockerfile-builder \
    . -t puff-builder
docker run -v "./target:/output" --rm puff-builder

echo "-----------------"
echo "Image (using version label: $PVER)"
echo "-----------------"
docker build \
    --build-arg PUFF_CURRENT_VERSION=$PVER \
    -f Dockerfile \
    ./target -t ghcr.io/lukfisz/puff:latest

echo "-----------------"
echo "VHS"
echo "-----------------"
vhs preview.tape

echo "-----------------"
echo "Cleaning"
echo "-----------------"
rm -f ./target/puff_*
