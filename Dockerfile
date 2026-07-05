# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS backend-builder

ARG APP_CMD=api
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${APP_CMD}

FROM alpine:3.20 AS backend

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=backend-builder /out/app /app/app

RUN mkdir -p /app/out/logs && chown -R app:app /app
USER app

EXPOSE 8080
ENTRYPOINT ["/app/app"]

FROM node:22-bookworm-slim AS frontend-builder

WORKDIR /app

COPY frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY frontend/ ./
RUN npm run build

FROM nginx:1.27-alpine AS frontend

COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
