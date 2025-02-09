FROM golang:1.20-alpine

# working directory
WORKDIR /app

# Copy code and install dependencies
COPY . .
RUN go mod download

# Build the Go application
RUN go build -o main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
