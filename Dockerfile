# Build stage
FROM golang:alpine as Builder

WORKDIR /app

COPY go.mod go.sum ./

COPY . .

RUN go build -o main .

# Execute stage
FROM alpine:latest

WORKDIR /root/

COPY --from=Builder /app/main .

CMD ["./main"] # Execute
