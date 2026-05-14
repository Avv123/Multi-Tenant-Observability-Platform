FROM node:22-alpine AS builder
WORKDIR /app

COPY services/pulselens-ui/package.json services/pulselens-ui/package-lock.json ./
RUN npm ci
COPY services/pulselens-ui ./
RUN npm run build

FROM node:22-alpine
WORKDIR /app
RUN npm install -g serve
COPY --from=builder /app/dist ./dist
EXPOSE 3000

CMD ["serve", "-s", "dist", "-l", "3000"]
