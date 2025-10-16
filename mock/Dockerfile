# =================== Build stage ===================
FROM golang:1.21-alpine AS builder

# Diretório de trabalho
WORKDIR /app

# Dependências do sistema
RUN apk add --no-cache git

# Copia arquivos de dependências Go e baixa módulos
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código
COPY . ./

# Compila o binário
RUN go build -o mock_adapter ./main.go

# =================== Runtime stage ===================
FROM alpine:latest

WORKDIR /app

# Copia o binário compilado do builder
COPY --from=builder /app/mock_adapter ./

# Configuração padrão via variáveis de ambiente (podem ser sobrescritas)
ENV DEVICE_API_URL=http://host.docker.internal:8080/devices
ENV RABBITMQ_URL=amqp://guest:guest@host.docker.internal:5672/
ENV POLL_MS=1000
ENV QUEUE_NAME=device.updates
ENV DELTA_RANGE=0.05
ENV RANDOM_SEED=0

# Expõe porta se necessário (não obrigatório para mock)
EXPOSE 8080

# Comando padrão
CMD ["./mock_adapter"]
