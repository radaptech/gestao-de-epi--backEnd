# ==========================================
# ESTÁGIO 1: O Construtor (Builder)
# ==========================================
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Copia e baixa as dependências primeiro (O cache do Docker agradece!)
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o resto do seu código
COPY . .

# Compila o binário otimizado
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# ==========================================
# ESTÁGIO 2: A Imagem Final (Produção)
# ==========================================
FROM alpine:latest

WORKDIR /app

# Instala certificados e configura fuso horário
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Sao_Paulo

# Desliga os logs coloridos de debug do GIN para ganhar performance
ENV GIN_MODE=release

# 👇 A MÁGICA DOS DIRETÓRIOS AQUI:
# Mantemos a mesma estrutura de pastas do seu projeto original para o Go não se perder
COPY --from=builder /app/database/migrate ./database/migrate

# Copia o binário já compilado
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]