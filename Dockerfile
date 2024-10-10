# Use an official Golang image as the base image
FROM golang:1.23.2-alpine as build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go app
RUN go build -o crypto-wallet-app

# Expose the application port
EXPOSE 8085

# Run the Go app
CMD ["./crypto-wallet-app"]
