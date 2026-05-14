package helper

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

func GerarTokenAuditoria(nome, funcao, depto string, data time.Time) string {
    // 1. Monta a string base com os dados (repare no formato de data fixo para o hash)
    payload := fmt.Sprintf("%s|%s|%s|%s", 
        strings.ToUpper(nome), 
        strings.ToUpper(funcao), 
        strings.ToUpper(depto), 
        data.Format("2006-01-02"),
    )

    // 2. Gera o Hash SHA-256
    hash := sha256.Sum256([]byte(payload))

    // 3. Retorna apenas os primeiros 12 ou 16 caracteres em maiúsculo para ficar elegante
    // Exemplo: ENT-A1B2C3D4E5F6
    return fmt.Sprintf("ENT-%X", hash)[:16]
}

func GerarTokenDevolucao(nome, funcao, depto string, data time.Time) string {
    // 1. Monta a string base com os dados (repare no formato de data fixo para o hash)
    payload := fmt.Sprintf("%s|%s|%s|%s", 
        strings.ToUpper(nome), 
        strings.ToUpper(funcao), 
        strings.ToUpper(depto), 
        data.Format("2006-01-02"),
    )

    // 2. Gera o Hash SHA-256
    hash := sha256.Sum256([]byte(payload))

    // 3. Retorna apenas os primeiros 12 ou 16 caracteres em maiúsculo para ficar elegante
    // Exemplo: ENT-A1B2C3D4E5F6
    return fmt.Sprintf("DEVO-%X", hash)[:16]
}