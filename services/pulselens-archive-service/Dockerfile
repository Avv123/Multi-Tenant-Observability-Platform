FROM golang:1.24 AS builder
WORKDIR /workspace

COPY libs ./libs
COPY services/pulselens-archive-service ./services/pulselens-archive-service

WORKDIR /workspace/services/pulselens-archive-service
RUN GOWORK=off go build -o /out/pulselens-archive-service .

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/pulselens-archive-service /app/pulselens-archive-service
COPY services/pulselens-archive-service/configs/docker.yaml /app/configs/config.yaml

ENV CONFIG_PATH=/app/configs/config.yaml
EXPOSE 8086

CMD ["./pulselens-archive-service"]
