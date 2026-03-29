# ==========================================
# ESTÁGIO 1: O Construtor (Builder)
# ==========================================
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Copia e baixa as dependências primeiro (O Docker faz cache disso e acelera muito os próximos builds)
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o resto do seu código
COPY . .

# Compila o binário. O comando -ldflags="-w -s" remove informações de debug e deixa o arquivo incrivelmente menor!
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main main.go

# ==========================================
# ESTÁGIO 2: A Imagem Final (Produção)
# ==========================================
# Aqui a gente pega um Linux Alpine "limpo", que pesa apenas 5 Megabytes!
FROM alpine:latest

WORKDIR /app

# Instala apenas os certificados de segurança e o fuso horário brasileiro
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Sao_Paulo

# A MÁGICA ACONTECE AQUI:
# Nós roubamos APENAS o arquivo 'main' (que já está compilado) lá do estágio 1.
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]