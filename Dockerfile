
# Start from the official Golang image
FROM golang:1.17

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o edith -ldflags="-X 'github.com/theycallmeloki/edith-cli/cmd/edithctl.version=0.0.1'" main.go

# Copy the Edith configuration file into the container
COPY edith.json /root/.config/edith/edith.json


# Expose the port the app runs on (not required yet)
# EXPOSE 8080

# Run the app
CMD ["./edith", "--help"]