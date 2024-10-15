# Crypto Wallet Application

This project consists of a React frontend and a Go backend for a crypto wallets management application.

## Project Structure

```
.
├── frontend/       # React application
│   └── .env.example
|   └── Dockerfile  # Dockerfile for building the frontend
|   └── *           # React source files
├── backend/        # Go application
│   └── .env.example
|   └── Dockerfile  # Dockerfile for building the backend
|   └── *           # Go source files
└── docker-compose.yml
```

## Prerequisites

- Docker
- Docker Compose

## Setup

1. Clone the repository:
   ```
   git clone <repository-url>
   cd <project-directory>
   ```

2. Set up environment variables:
   - In the `frontend/` directory, copy `.env.example` to `.env` and adjust the variables as needed.
   - In the `backend/` directory, copy `.env.example` to `.env` and adjust the variables as needed.

## Running the Application

To run the entire application using Docker Compose:

1. Build and start the containers:
   ```
   docker-compose up --build
   ```

2. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8085

To stop the application:

```
docker-compose down
```

## Docker Compose

The `docker-compose.yml` file defines the services for frontend, backend and database. It sets up the necessary environment and connections between the services.

To modify the Docker Compose configuration, edit the `docker-compose.yml` file in the root directory.

## Structure
### Backend
The backend is a Go application that provides a RESTful API for managing crypto wallets. It uses the Gin framework for routing and middleware. 
The application is structured with the following layers and services:
- **Handlers**: Handle incoming HTTP requests and return responses by interacting with the services.
- **Services**: Implement the business logic and interact with the repository layer, KMS and Web3 services.
- **Repositories**: Handle database operations.
- **Database**: Uses a MongoDB database to store user, wallet and transaction data.
- **Middleware**: Provides middleware functions for authentication and error handling.
- **KMS**: Provides a key management service (KMS) for managing wallet private keys. The private keys never touch the backend server; they are created and used to sign transactions directly via the KMS API.
- **Web3**: Provides a web3 service for interacting with the Ethereum blockchain.

## CI/CD Pipeline
The project uses Gitlab CI/CD for automating the build and deployment process. The `.gitlab-ci.yml` file defines the stages and jobs for the pipeline. The pipeline is triggered on every push to the repository. The pipeline consists of the following stages:
- **Build**: Builds the frontend and backend Docker images and pushes them to Docker Hub.
- **Deploy**: Deploys the application using Kubernetes. The deployment consists of a frontend, a backend and a database service, each with a single replica.

## Kubernetes Deployment
The application is configured to be deployed on a Kubernetes cluster. The `k8s/` directory contains the Kubernetes manifests for deploying the frontend, backend and database services. The deployment consists of the following components:
- **Frontend Service**: Exposes the frontend application to the internet.
- **Backend Service**: Exposes the backend API to the internet.
- **MongoDB StatefulSet**: Deploys a MongoDB database to store user, wallet and transaction data, which is exposed as a service for the backend to connect to.


## Configuration

Adjust the environment variables in the `.env` files for both frontend and backend to configure the application. Refer to the `.env.example` files for the required variables. In case of kubernetes deployment, the environment variables should be set in the gitlab CI/CD pipeline.