package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedEmpresaMatriz(db *pgxpool.Pool) error {
	ctx := context.Background()
	log.Println("Verificando Empresa Matriz e Super Admin...")

	var empresaID int64

	// 1. Tenta buscar a empresa primeiro para ver se já existe (Idempotência)
	err := db.QueryRow(ctx, `SELECT id FROM empresas WHERE nome_fantasia = $1`, "sge-gestaoEpi").Scan(&empresaID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 2. A empresa não existe! Vamos INSERIR e JÁ PEGAR o ID na mesma linha com RETURNING
			log.Println("Empresa Matriz não encontrada. Criando...")
			
			err = db.QueryRow(ctx, `
				INSERT INTO empresas (nome_fantasia, razao_social, cnpj, subdominio) 
				VALUES ('sge-gestaoEpi', 'radaptech', '53563447', 'painel-homologacao')
				RETURNING id;
			`).Scan(&empresaID) // Pega o ID devolvido pelo banco instantaneamente

			if err != nil {
				return fmt.Errorf("falha ao criar empresa matriz: %v", err)
			}
		} else {
			return fmt.Errorf("erro ao verificar empresa no banco: %v", err)
		}
	} else {
		log.Println("Empresa Matriz já existe. Pulando criação.")
	}

	// ==========================================
	// CONFIGURAÇÃO DO USUÁRIO ADMIN
	// ==========================================

	emailAdmin := os.Getenv("SUPER_ADMIN_EMAIL")
	if emailAdmin == "" {
		emailAdmin = "admin@radaptech.com.br"
	}

	// 3. Verifica se o usuário já existe antes de tentar criar de novo
	var adminID int64
	err = db.QueryRow(ctx, `SELECT id FROM usuarios WHERE email = $1`, emailAdmin).Scan(&adminID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("Super Admin não encontrado. Criando...")

			senhaPadrao := os.Getenv("SUPER_ADMIN_PASSWORD")
			if senhaPadrao == "" {
				senhaPadrao = "Mudar@123"
			}

			// 4. Tratando o erro do Hash (Nunca use o _ para ignorar erros de segurança)
			hash, errHash := auth.HashPassword(senhaPadrao)
			if errHash != nil {
				return fmt.Errorf("falha ao gerar hash da senha: %v", errHash)
			}

			// 5. Insere o Admin vinculando à empresaID garantida
			_, err = db.Exec(ctx, `
				INSERT INTO usuarios (tenant_id, nome, email, senha_hash, ativo, role)
				VALUES ($1, 'Administrador SGE', $2, $3, true, 'super_admin')
			`, empresaID, emailAdmin, string(hash))

			if err != nil {
				return fmt.Errorf("falha ao criar super admin: %v", err)
			}
			
			log.Println("Super Admin criado com sucesso!")
		} else {
			return fmt.Errorf("erro ao verificar usuario admin no banco: %v", err)
		}
	} else {
		log.Println("Super Admin já existe. Pulando criação.")
	}

	return nil
}
