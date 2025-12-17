FROM golang:1.24.2

WORKDIR /app

# Install protoc and dependencies
RUN apt-get update && apt-get install -y protobuf-compiler

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .


RUN go build -o main .

EXPOSE 50051

CMD ["./main"]