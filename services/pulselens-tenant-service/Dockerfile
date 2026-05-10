FROM golang:1.24 AS builder
WORKDIR /workspace

COPY libs ./libs
COPY services/pulselens-tenant-service ./services/pulselens-tenant-service

WORKDIR /workspace/services/pulselens-tenant-service
RUN GOWORK=off go build -o /out/pulselens-tenant-service .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /out/pulselens-tenant-service /app/pulselens-tenant-service
COPY services/pulselens-tenant-service/configs/docker.yaml /app/configs/config.yaml

ENV CONFIG_PATH=/app/configs/config.yaml
EXPOSE 8081

CMD ["./pulselens-tenant-service"]
