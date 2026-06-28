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
	var planoid int64

	// ==========================================
	// 1. CONFIGURAÇÃO DO PLANO MATRIZ
	// ==========================================
	errPlano := db.QueryRow(ctx, `SELECT id FROM planos WHERE nome = $1`, "plano de testes").Scan(&planoid)
	
	if errPlano != nil {
		if errors.Is(errPlano, pgx.ErrNoRows) {
			log.Println("Plano Matriz não encontrado. Criando...")

			
			
			errPlano = db.QueryRow(ctx, `
				INSERT INTO planos (nome, mensalidade, limite_funcionarios, limite_usuarios, limite_epis, status, descricao, criado_em)
				VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
				RETURNING id;
			`, "plano de testes", 1000.00, 10000, 10000, 1000, "Ativo", "plano de testes").Scan(&planoid)

			if errPlano != nil {
				return fmt.Errorf("falha ao criar plano matriz: %v", errPlano)
			}
		} else {
			return fmt.Errorf("erro ao verificar plano no banco: %v", errPlano)
		}
	} else {
		log.Println("Plano Matriz já existe. Pulando criação.")
	}

	// ==========================================
	// 2. CONFIGURAÇÃO DA EMPRESA MATRIZ
	// ==========================================
	err := db.QueryRow(ctx, `SELECT id FROM empresas WHERE nome_fantasia = $1`, "sge-gestaoEpi").Scan(&empresaID)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("Empresa Matriz não encontrada. Criando...")

			
			err = db.QueryRow(ctx, `
				INSERT INTO empresas (nome_fantasia, razao_social, cnpj, subdominio, criado_em, plano_id, status, vencimento, observacoes, responsavel, email, telefone) 
				VALUES ($1, $2, $3, $4, NOW(), $5, $6, $7, $8, $9, $10, $11)
				RETURNING id;
			`, 
				"sge-gestaoEpi", "radaptech", "53563447", "painel-homologacao", 
				planoid, "Ativa", "2099-03-14", "testes", "radaptech", 
				"radaptech@gmail.com", "34998972788",
			).Scan(&empresaID)

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
	// 3. CONFIGURAÇÃO DO USUÁRIO ADMIN
	// ==========================================
	emailAdmin := os.Getenv("SUPER_ADMIN_EMAIL")
	if emailAdmin == "" {
		emailAdmin = "admin@radaptech.com.br"
	}

	var adminID int64
	err = db.QueryRow(ctx, `SELECT id FROM usuarios WHERE email = $1`, emailAdmin).Scan(&adminID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("Super Admin não encontrado. Criando...")

			senhaPadrao := os.Getenv("SUPER_ADMIN_PASSWORD")
			if senhaPadrao == "" {
				senhaPadrao = "Mudar@123"
			}

			hash, errHash := auth.HashPassword(senhaPadrao)
			if errHash != nil {
				return fmt.Errorf("falha ao gerar hash da senha: %v", errHash)
			}

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
