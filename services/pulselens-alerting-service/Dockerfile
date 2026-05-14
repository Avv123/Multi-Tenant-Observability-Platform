FROM golang:1.24 AS builder
WORKDIR /workspace

COPY libs ./libs
COPY services/pulselens-alerting-service ./services/pulselens-alerting-service

WORKDIR /workspace/services/pulselens-alerting-service
RUN GOWORK=off go build -o /out/pulselens-alerting-service .

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/pulselens-alerting-service /app/pulselens-alerting-service
COPY services/pulselens-alerting-service/configs/docker.yaml /app/configs/config.yaml

ENV CONFIG_PATH=/app/configs/config.yaml
EXPOSE 8085

CMD ["./pulselens-alerting-service"]
