# Build stage
FROM golang:1.22.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o wikigraph_server .


FROM alpine:latest
WORKDIR /app

# Copy Env
COPY .env .
# Copy graph data
COPY wikipedia_binary_graph.bin .
# Copy only the binary from the build stage
COPY --from=builder /app/wikigraph_server .
EXPOSE 8080

CMD ["./wikigraph_server"]
