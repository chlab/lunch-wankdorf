#!/usr/bin/sh

set -e

docker buildx build --platform linux/amd64 -t chlab/lunch-wankdorf:latest ../
docker push chlab/lunch-wankdorf:latest