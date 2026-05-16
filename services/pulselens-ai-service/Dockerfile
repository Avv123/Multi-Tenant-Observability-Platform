FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ai-service .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates curl
WORKDIR /app
COPY --from=builder /app/ai-service .
EXPOSE 8085
CMD ["./ai-service"]
