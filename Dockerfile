# 1. Ajuste a versão para uma existente (1.23 ou 1.22)
FROM golang:1.26.1-alpine

WORKDIR /app

# Instala dependências
RUN apk --no-cache add ca-certificates tzdata

ENV TZ=America/Sao_Paulo

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 2. O PULO DO GATO AQUI:
# Forçamos o build para Linux e arquitetura correta, ignorando o que veio do seu PC
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

EXPOSE 8080

CMD ["./main"]