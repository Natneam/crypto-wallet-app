#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Applying Kubernetes manifests..."

# Apply MongoDB deployment and service
echo "Applying MongoDB deployment and service..."
kubectl apply -f mongodb-deployment.yaml

# Wait for MongoDB pod to be ready
echo "Waiting for MongoDB pod to be ready..."
kubectl wait --for=condition=ready pod -l app=mongodb --timeout=300s

# Apply Backend deployment and service
echo "Applying Backend deployment and service..."
kubectl apply -f backend-deployment.yaml

# Wait for Backend pod to be ready
echo "Waiting for Backend pod to be ready..."
kubectl wait --for=condition=ready pod -l app=backend --timeout=300s

# Apply Frontend deployment and service
echo "Applying Frontend deployment and service..."
kubectl apply -f frontend-deployment.yaml

# Wait for Frontend pod to be ready
echo "Waiting for Frontend pod to be ready..."
kubectl wait --for=condition=ready pod -l app=frontend --timeout=300s

echo "All manifests applied successfully!"

# Display pod status
echo "Current pod status:"
kubectl get pods

# Display services
echo "Available services:"
kubectl get services

# Port forward frontend service

# Port forward frontend service
nohup kubectl port-forward svc/frontend 3000:3000 > frontend.log 2>&1 &

# Port forward backend service
nohup kubectl port-forward svc/backend 8085:8085 > backend.log 2>&1 &

echo "Setup complete! You can now access your application."