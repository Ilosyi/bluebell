# syntax=docker/dockerfile:1

FROM golang:1.26.1-alpine AS backend-builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bluebell .

FROM alpine:3.20 AS backend

RUN apk add --no-cache ca-certificates tzdata \
  && addgroup -S bluebell \
  && adduser -S -G bluebell bluebell \
  && mkdir -p /app/settings /app/logs \
  && chown -R bluebell:bluebell /app

WORKDIR /app

COPY --from=backend-builder /out/bluebell /app/bluebell
COPY settings/config.docker.yaml /app/settings/config.docker.yaml

ENV BLUEBELL_CONFIG_FILE=/app/settings/config.docker.yaml
ENV TZ=Asia/Shanghai

USER bluebell

EXPOSE 8080

ENTRYPOINT ["/app/bluebell"]

FROM node:20-alpine AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM nginx:1.27-alpine AS frontend

COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-builder /src/frontend/dist /usr/share/nginx/html

EXPOSE 80
