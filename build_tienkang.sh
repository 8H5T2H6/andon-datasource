#!/bin/bash

# 正式用
VERSION="tienkang"
CONTAINER="andon-datasource"

# docker location
DOCKER_REPO="iiicondor/$CONTAINER"

docker build -t $DOCKER_REPO:$VERSION .
# docker push $DOCKER_REPO:$VERSION
docker push $DOCKER_REPO:$VERSION

# docker repo 
# docker tag $DOCKER_REPO:$VERSION $DOCKER_REPO:latest
# docker push $DOCKER_REPO:latest

# harbor repo
# docker tag $DOCKER_REPO:$VERSION $HARBOR_REPO:latest
# docker push $HARBOR_REPO:latest

# docker rmi -f $(docker images | grep $CONTAINER | awk '{print $3}')