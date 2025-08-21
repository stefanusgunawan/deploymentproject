# Use official Golang image
FROM golang:1.22

# Set working directory inside container
WORKDIR /app

# Copy all files into container
COPY . .

# Build the Go app into a binary called "server"
RUN go build -o server .

# Run the binary
CMD ["./server"]