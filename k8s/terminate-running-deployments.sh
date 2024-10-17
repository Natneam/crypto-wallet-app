#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Terminating running deployments..."
kubectl delete deployment frontend backend mongodb