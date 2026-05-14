FROM golang:1.24 AS builder
WORKDIR /workspace

COPY libs ./libs
COPY services/pulselens-ingest-service ./services/pulselens-ingest-service

WORKDIR /workspace/services/pulselens-ingest-service
RUN GOWORK=off go build -o /out/pulselens-ingest-service .

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/pulselens-ingest-service /app/pulselens-ingest-service
COPY services/pulselens-ingest-service/configs/docker.yaml /app/configs/config.yaml

ENV CONFIG_PATH=/app/configs/config.yaml
EXPOSE 8082

CMD ["./pulselens-ingest-service"]
