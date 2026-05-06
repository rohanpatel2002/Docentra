FROM golang:1.22-alpine

WORKDIR /app

# We will copy go.mod and go.sum once they are initialized
COPY . .

# Expose port and run a placeholder command until we compile the Go app
EXPOSE 8080
CMD ["tail", "-f", "/dev/null"]
