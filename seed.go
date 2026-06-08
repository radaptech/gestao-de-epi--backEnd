package main

import (
	"context"
	"log"
	"os"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedEmpresaMatriz(db *pgxpool.Pool) error {
	ctx := context.Background()

	// 1. Verifica se a empresa matriz já existe
	var existe bool
	err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM empresas WHERE id = 1)").Scan(&existe)
	if err != nil {
		return err
	}

	if existe {
		log.Println("Empresa matriz já configurada. Pulando seed.")
		return nil
	}

	log.Println("Configurando Empresa Matriz pela primeira vez...")

	// 2. Insere a Empresa
	_, err = db.Exec(ctx, `
       insert into empresas(id,nome_fantasia, razao_social, cnpj, subdominio) values(
				1, 'sge-gestaoEpi', 'radaptech','53563447', 'painel'
		);
    `)
	if err != nil {
		return err
	}

	// 3. Pega a senha do arquivo .env (que fica seguro no servidor/Docker)
	senhaPadrao := os.Getenv("SUPER_ADMIN_PASSWORD")
	if senhaPadrao == "" {
		senhaPadrao = "Mudar@123" // Fallback de segurança
	}

	// 4. Gera o hash bcrypt em tempo de execução
	hash, _ := auth.HashPassword(senhaPadrao)

	// 5. Insere o usuário
	emailAdmin := os.Getenv("SUPER_ADMIN_EMAIL")
	if emailAdmin == "" {
    emailAdmin = "admin@radaptech.com.br" // Email padrão de segurança
    }
	_, err = db.Exec(ctx, `
        INSERT INTO usuarios (tenant_id, nome, email, senha_hash, ativo, role)
        VALUES (1, 'Administrador SGE', $1, $2, true, 'super_admin')
    `, emailAdmin, string(hash))

	if err != nil {
		return err
	}

	log.Println("Empresa Matriz e Super Admin criados com sucesso!")
	return nil
}
