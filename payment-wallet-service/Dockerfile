FROM golang:1.25-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application from cmd directory
RUN go build -o payment-processor-service ./cmd/

# Run the application
CMD ["./payment-processor-service"]