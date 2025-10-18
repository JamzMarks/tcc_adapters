# Etapa 1: build
FROM golang:1.25 AS builder

WORKDIR /app
# Copia os arquivos de dependências primeiro
COPY go.mod go.sum ./
RUN go mod download

# Copia o resto do código
COPY . .

# Compila o binário estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o adapter .
# Etapa 2: imagem final mínima
FROM alpine:latest

WORKDIR /app

# Copia o binário do builder
COPY --from=builder /app/adapter .

# Configuração padrão via variáveis de ambiente (podem ser sobrescritas)
ENV DEVICE_API_URL=http://host.docker.internal:8080/devices
ENV RABBITMQ_URL=amqp://guest:guest@host.docker.internal:5672/
ENV POLL_MS=1000
ENV QUEUE_NAME=injector_queue
ENV DELTA_RANGE=0.05
ENV RANDOM_SEED=0

# Expõe a porta 
EXPOSE 8080

# Comando padrão
CMD ["./adapter"]
