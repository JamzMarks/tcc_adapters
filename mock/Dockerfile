# ===========================
# Etapa 1: Build da aplicação
# ===========================
FROM golang:1.25.2-alpine AS builder

# Instalar dependências úteis
RUN apk add --no-cache git

# Diretório de trabalho dentro do container
WORKDIR /app

# Copiar os arquivos go.mod e go.sum antes para cachear dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o restante do código-fonte
COPY . .

# Compilar o binário
RUN go build -o mock .

# ===========================
# Etapa 2: Runtime (imagem final)
# ===========================
FROM alpine:3.20

WORKDIR /app

# Copia o binário do builder
COPY --from=builder /app/mock .

# Expõe a porta 
EXPOSE 8081

# Comando padrão
CMD ["./mock"]

