#!/bin/bash

echo "Starting Billing Note Application..."

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
  echo "Error: Docker is not running. Please start Docker and try again."
  exit 1
fi

echo "Building and starting containers..."
docker-compose up -d --build

echo ""
echo "============================================"
echo "Billing Note is starting up!"
echo "Frontend will be available at: http://localhost"
echo "Backend API will be available at: http://localhost:8080"
echo "============================================"
echo ""
echo "To stop the application, run: docker-compose down"
