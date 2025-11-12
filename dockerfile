# ---------------------
# 1️⃣ Etapa de Build
# ---------------------
FROM golang:1.23-alpine AS builder

# Instala git e certificados
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copia os arquivos de dependência e baixa módulos
COPY go.mod go.sum ./
RUN go mod download

# Copia TODO o código fonte (inclui db/migrations)
COPY . .

# Compila o binário
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ---------------------
# 2️⃣ Etapa Final
# ---------------------
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia o binário compilado
COPY --from=builder /app/main .

# ⚠️ Copia as migrations para o container final
COPY --from=builder /app/db ./db

EXPOSE 8080

CMD ["./main"]
