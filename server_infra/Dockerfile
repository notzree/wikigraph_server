# Use the official Golang image as the base image
FROM golang:1.22.3-alpine

# Install reflex for hot reloading
RUN apk add --no-cache git
RUN go install github.com/cespare/reflex@latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Expose the application on port 8080
EXPOSE 8080

# Command to run reflex for hot reloading
CMD ["reflex", "-r", "\\.go$", "-s", "--", "go", "run", ".", "-race"]
