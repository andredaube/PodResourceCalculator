#!/bin/sh

## build the PodResourceCalculator util and copy it out of the container

IMAGE_NAME=podresourcecalculator
TAG=1

docker build . -t ${IMAGE_NAME}:${TAG}

docker run --rm -v ./out:/out --name ${IMAGE_NAME} ${IMAGE_NAME}:${TAG} cp /PodResourceCalculator /out

docker rmi ${IMAGE_NAME}:${TAG}

