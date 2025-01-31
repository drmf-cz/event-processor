FROM golang:1.23

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Install mockery
RUN GOBIN=/usr/local/bin go install github.com/vektra/mockery/v2@v2.38.0

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Default command (can be overridden in docker-compose)
CMD ["go", "run", "main.go"] 