#!/bin/zsh

set -e

docker pull ghcr.io/lukfisz/puff:latest

vhs preview.tape
