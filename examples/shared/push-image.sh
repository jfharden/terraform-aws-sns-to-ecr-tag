#!/bin/bash

if [ "$#" -ne 4 ]; then
  echo "Usage: $0 <account_id> <region> <repo_name> <tag>"
  echo "example: $0 12345678 eu-west-2 hello-world-example latest"
  exit 2
fi

ACCOUNT_ID=$1
REGION=$2
REPO=$3
TAG=$4

ECR_ENDPOINT="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com"
REPO_URI="$ECR_ENDPOINT/$REPO"
aws ecr get-login-password --region "$REGION" | docker login --username AWS --password-stdin "$ECR_ENDPOINT"

docker pull hello-world:linux
docker tag hello-world:linux "$REPO_URI:$TAG"
docker push "$REPO_URI:$TAG"
