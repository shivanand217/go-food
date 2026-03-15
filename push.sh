#!/bin/bash

# ==============================================================================
# Go-Food - Docker Push Script
# This script tags and pushes the locally built microservice images to Docker Hub.
#
# Prerequisite: You must run `docker login` before executing this script.
# Use: ./push.sh <your_docker_hub_username>
# ==============================================================================

if [ -z "$1" ]; then
    echo "Error: Please provide your Docker Hub username."
    echo "Usage: ./push.sh <your_docker_hub_username>"
    exit 1
fi

USERNAME=$1
VERSION="v1.0"

echo "Using Docker Hub username: $USERNAME"

# 1. Build the images first just to be sure
echo "Building latest images..."
docker-compose build

# List of all our custom microservices
SERVICES=("api-gateway" "user-service" "restaurant-service" "order-service" "delivery-service")

echo "Tagging and pushing images to Docker Hub..."

for SERVICE in "${SERVICES[@]}"; do
    # docker-compose generates local images named like 'go-food-api-gateway'
    # We strip 'go-food-' or rely on the directory name. 
    # By default, docker compose names images: <project_folder_name>-<service_name>
    LOCAL_IMAGE="go-food-$SERVICE"
    REMOTE_IMAGE="$USERNAME/go-food-$SERVICE:$VERSION"
    
    echo "----------------------------------------"
    echo "Preparing $SERVICE..."
    
    # Check if local image exists
    if [[ "$(docker images -q $LOCAL_IMAGE 2> /dev/null)" == "" ]]; then
      echo "Warning: Local image $LOCAL_IMAGE not found. Skipping..."
      continue
    fi

    # Tag the local image for the remote repository
    docker tag $LOCAL_IMAGE $REMOTE_IMAGE
    echo "Tagged as $REMOTE_IMAGE"

    # Push to Docker Hub
    echo "Pushing $REMOTE_IMAGE..."
    docker push $REMOTE_IMAGE
done

echo "----------------------------------------"
echo "✅ All images pushed successfully!"
echo "Anyone can now run your microservices by pulling $USERNAME/go-food-<service_name>:$VERSION"
