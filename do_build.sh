#!/bin/sh

## build the PodResourceCalculator util and copy it out of the container

IMAGE_NAME=podresourcecalculator
TAG=1

UID=$(id -u)
GID=$(id -g)
IDUN=$(id -un)

docker build . -t ${IMAGE_NAME}:${TAG} \
  --build-arg UID=${UID} \
  --build-arg GID=${GID} \
  --build-arg IDUN=${IDUN}

docker run --rm -v ./:/out \
  --name ${IMAGE_NAME} \
  ${IMAGE_NAME}:${TAG} \
  cp /src/PodResourceCalculator /out

if [ $? -eq 0 ]; then
  type strip > /dev/null 2>&1 && strip ./PodResourceCalculator
fi

