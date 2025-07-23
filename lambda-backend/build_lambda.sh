#!/bin/bash

set -e  # Exit on any error

# Optional: customize Dockerfile name and output file
DOCKERFILE="Dockerfile.lambda"
IMAGE_NAME="lambda-builder"
ZIP_OUTPUT="function.zip"

echo "Building Docker image..."
docker build -f "$DOCKERFILE" -t "$IMAGE_NAME" .

echo "Creating container to extract zip..."
container_id=$(docker create "$IMAGE_NAME")

echo "Copying $ZIP_OUTPUT from container..."
docker cp "$container_id":/tmp/"$ZIP_OUTPUT" .

echo "Removing temp container..."
docker rm "$container_id" > /dev/null

echo "Lambda package built: $ZIP_OUTPUT"

