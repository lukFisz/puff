#!/bin/zsh

set -e

local output_path=$1

if [ -z "$output_path" ]; then
    echo "Specify output path:"
    echo "> ./build.sh <output path>"
    exit 1
fi


echo "-----------------------------------"
echo "Puff builder"
echo "-----------------------------------"
docker build \
    -f builder.Dockerfile \
    . -t ghcr.io/lukfisz/puff-builder
docker run -v "$output_path:/output" --rm ghcr.io/lukfisz/puff-builder
