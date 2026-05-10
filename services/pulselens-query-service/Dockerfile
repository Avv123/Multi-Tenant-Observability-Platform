FROM node:22-alpine AS ui-builder
WORKDIR /workspace/services/pulselens-ui
COPY services/pulselens-ui/package.json services/pulselens-ui/package-lock.json ./
RUN npm ci
COPY services/pulselens-ui ./
RUN npm run build

FROM golang:1.24 AS go-builder
WORKDIR /workspace

COPY libs ./libs
COPY services/pulselens-query-service ./services/pulselens-query-service

WORKDIR /workspace/services/pulselens-query-service
RUN GOWORK=off go build -o /out/pulselens-query-service .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=go-builder /out/pulselens-query-service /app/pulselens-query-service
COPY services/pulselens-query-service/configs/docker.yaml /app/configs/config.yaml
COPY --from=ui-builder /workspace/services/pulselens-ui/dist /app/ui

ENV CONFIG_PATH=/app/configs/config.yaml
EXPOSE 8084

CMD ["./pulselens-query-service"]
