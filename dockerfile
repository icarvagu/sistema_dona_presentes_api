# ---------------------
# 1️⃣ Etapa de Build
# ---------------------
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ---------------------
# 2️⃣ Etapa Final (com Chromium para PDF)
# ---------------------
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    chromium \
    fonts-liberation \
    && rm -rf /var/lib/apt/lists/*

# Chromium sem sandbox (necessário em container)
ENV CHROME_BIN=/usr/bin/chromium
ENV CHROMEDP_SKIP_CHROMIUM_DOWNLOAD=1
ENV CHROMEDP_HEADLESS=1

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/db ./db
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./main"]
