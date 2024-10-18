# Use golang:1.23-alpine as the base image
FROM golang:1.23-alpine

# Set the working directory in the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Copy the lib directory
COPY lib ./lib

# Copy the authentication service files
COPY services/notification ./services/notification

# Copy the .env file
COPY .env ./

# Download dependencies
RUN go mod download

# Build the authentication service
RUN go build -o notification ./services/notification

# Expose the port the service runs on
EXPOSE 8080

# Command to run the service
CMD ["./notification"]
