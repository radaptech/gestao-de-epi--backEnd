package service

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// --- UTILITÁRIOS ---

func randomString(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func randomInt() *rand.Rand {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return r
}

// Helper para gerar um "CNPJ" numérico aleatório de 14 dígitos
func randomCNPJ() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%014d", rand.Int63n(99999999999999))
}
// --- 1. EMPRESA (TENANT) ---

func CreateEmpresa(t *testing.T, db *pgxpool.Pool) int64 {
	var id int64
	query := `
		INSERT INTO empresas (nome_fantasia, razao_social, cnpj, subdominio, ativo) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id;
	`
	// Gera dados aleatórios para satisfazer as constraints UNIQUE
	nano := time.Now().UnixNano()
	nomeFantasia := fmt.Sprintf("Radap Client %d", nano)
	razaoSocial := fmt.Sprintf("Radap Tech Clientes Ltda %d", nano)
	cnpj := fmt.Sprintf("%d", nano)       // CNPJ único (fake)
	subdominio := fmt.Sprintf("cliente%d", nano) // Subdominio único

	err := db.QueryRow(context.Background(), query, 
		nomeFantasia, 
		razaoSocial, 
		cnpj, 
		subdominio, 
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEmpresa falhou: %v", err)
	}
	return id
}

// --- 2. USUÁRIO (Vinculado ao Tenant) ---

func CreateUser(t *testing.T, db *pgxpool.Pool, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO usuarios (tenant_id, nome, email, senha_hash, ativo) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	nome := randomString("Usuario Teste")
	// O email deve ser único por tenant
	email := fmt.Sprintf("user_%d@radap.com", time.Now().UnixNano())
	senha := "hash_senha_segura"

	err := db.QueryRow(context.Background(), query,
		tenantID,
		nome,
		email,
		senha,
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateUser falhou: %v", err)
	}
	return id
}

// --- 3. ESTRUTURA ORGANIZACIONAL ---

func CreateDepartamento(t *testing.T, db *pgxpool.Pool, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO departamento (tenant_id, nome, ativo) 
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query, 
		tenantID, 
		randomString("Dep"), 
		true,
	).Scan(&id)
	
	if err != nil {
		t.Fatalf("Helper CreateDepartamento falhou: %v", err)
	}
	return id
}

func CreateFuncao(t *testing.T, db *pgxpool.Pool, idDep, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO funcao (tenant_id, nome, IdDepartamento, ativo) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query, 
		tenantID, 
		randomString("Func"), 
		idDep, 
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateFuncao falhou: %v", err)
	}
	return id
}

func CreateFuncionario(t *testing.T, db *pgxpool.Pool, IdDepartamento, IdFuncao, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO funcionario (tenant_id, nome, IdFuncao, IdDepartamento, ativo) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`
	
	nome := randomString("Funcionario")
	

	err := db.QueryRow(context.Background(), query,
		tenantID,
		nome,
		IdFuncao,
		IdDepartamento,
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateFuncionario falhou: %v", err)
	}
	return id
}

// --- 4. CATALOGO DE EPIs ---

func CreateProtecao(t *testing.T, db *pgxpool.Pool, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO tipo_protecao (tenant_id, nome, ativo) 
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query, 
		tenantID, 
		randomString("Protecao"), 
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateProtecao falhou: %v", err)
	}
	return id
}

func CreateTamanho(t *testing.T, db *pgxpool.Pool, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO tamanho (tenant_id, tamanho, ativo) 
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query, 
		tenantID, 
		randomString("Tam"), 
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateTamanho falhou: %v", err)
	}
	return id
}

func CreateEpi(t *testing.T, db *pgxpool.Pool, idTipoProtecao, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO epi (
			tenant_id, nome, fabricante, CA, descricao, 
			validade_CA, IdTipoProtecao, alerta_minimo, ativo
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id;
	`
	r := randomInt()
	nome := randomString("Luva X")
	fabricante := randomString("Fabr Y")
	ca := fmt.Sprintf("%d", r.Intn(9999999)) // CA único por tenant
	descricao := "Descrição teste"
	validade := time.Now().AddDate(1, 0, 0)
	alertaMinimo := 10

	err := db.QueryRow(context.Background(), query,
		tenantID,
		nome,
		fabricante,
		ca,
		descricao,
		validade,
		idTipoProtecao,
		alertaMinimo,
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEpi falhou: %v", err)
	}
	return id
}

// --- 5. ESTOQUE E MOVIMENTAÇÃO ---

func CreateEntradaEpi(t *testing.T, db *pgxpool.Pool, IdFuncionario, idEpi, IdTipoProtecao, IdTamanho,Idfornecedor,idUserCriacao, tenantID int64) int64 {
	var id int64

	// Atenção: Adicionado id_usuario_criacao, nota_fiscal e tenant_id
	query := `
		INSERT INTO entrada_epi (
			tenant_id, IdEpi, IdTamanho, data_entrada, quantidade, quantidadeAtual, 
			data_fabricacao, data_validade, lote, Idfornecedor, valor_unitario, 
			nota_fiscal_numero, nota_fiscal_serie, id_usuario_criacao, ativo
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id;
	`

	lote := randomString("Lote")
	// Constraint: UNIQUE (nota_fiscal_numero, nota_fiscal_serie, fornecedor)
	// Usamos randomString para garantir unicidade nos testes
	
	notaFiscalNumero := randomString("NF")
	notaFiscalSerie := "1"

	err := db.QueryRow(context.Background(), query,
		tenantID,
		idEpi,
		IdTamanho,
		time.Now(), // Data Entrada (Tipo Date no banco, Driver converte)
		100,        // Quantidade Inicial
		100,        // Quantidade Atual
		time.Now().AddDate(-1, 0, 0), // Fabricação
		time.Now().AddDate(1, 0, 0),  // Validade
		lote,
		Idfornecedor,
		decimal.NewFromFloat(23.99),
		notaFiscalNumero,
		notaFiscalSerie,
		idUserCriacao,
		true, // Ativo
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEntradaEpi falhou: %v", err)
	}
	return id
}

// Versão com quantidade 1 para testes de limite
func CreateEntradaEpi1(t *testing.T, db *pgxpool.Pool, IdFuncionario, idEpi, IdTipoProtecao, IdTamanho, idUserCriacao, Idfornecedor,tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO entrada_epi (
			tenant_id, IdEpi, IdTamanho, data_entrada, quantidade, quantidadeAtual, 
			data_fabricacao, data_validade, lote, Idfornecedor, valor_unitario, 
			nota_fiscal_numero, nota_fiscal_serie, id_usuario_criacao, ativo
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id;
	`
	
	lote := randomString("Lote")
	
	notaFiscalNumero := randomString("NF")

	err := db.QueryRow(context.Background(), query,
		tenantID,
		idEpi,
		IdTamanho,
		time.Now(),
		1, // Quantidade 1
		1, // Quantidade Atual 1
		time.Now().AddDate(-1, 0, 0),
		time.Now().AddDate(1, 0, 0),
		lote,
		Idfornecedor,
		decimal.NewFromFloat(23.99),
		notaFiscalNumero,
		"1",
		idUserCriacao,
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEntradaEpi1 falhou: %v", err)
	}
	return id
}

func CreateEntregaEpi(t *testing.T, db *pgxpool.Pool, idFuncionario, idUserEntrega, tenantID int64) int64 {
	var id int64
	// Adicionado token_validacao e id_usuario_entrega
	query := `
		INSERT INTO entrega_epi (
			tenant_id, IdFuncionario, data_entrega, assinatura, 
			token_validacao, id_usuario_entrega, ativo
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`
	assinatura := randomString("AssinaturaBase64")
	token := randomString("TokenValidacao")

	err := db.QueryRow(context.Background(), query,
		tenantID,
		idFuncionario,
		time.Now(),
		assinatura,
		token,
		idUserEntrega,
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEntregaEpi falhou: %v", err)
	}
	return id
}

func CreateEpiEntregues(t *testing.T, db *pgxpool.Pool, idEntrega, idEntrada, IdEpi, IdTamanho, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO epis_entregues (
			tenant_id, IdEntrega, IdEntrada, IdEpi, IdTamanho, quantidade, ativo
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`
	quantidade := 10

	err := db.QueryRow(context.Background(), query,
		tenantID,
		idEntrega,
		idEntrada,
		IdEpi,
		IdTamanho,
		quantidade,
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEpiEntregues falhou: %v", err)
	}
	return id
}

func CreateMotivoDevolucao(t *testing.T, db *pgxpool.Pool, motivo string, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO motivo_devolucao (tenant_id, motivo, ativo) 
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query,
		tenantID,
		motivo,
		true,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateMotivoDevolucao falhou: %v", err)
	}
	return id
}

func CreateFornecedor(t *testing.T, db *pgxpool.Pool, tenantID int64) int64 {
	var id int64
	
	// Preenchemos nome_fantasia igual a razao_social e IE como 'ISENTO' para simplificar o helper
	query := `
		INSERT INTO fornecedores (
			tenant_id, 
			razao_social, 
			nome_fantasia, 
			cnpj, 
			inscricao_estadual, 
			ativo
		) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`
	razaoSocial:= randomString("razao")
	fantasia:= randomString("fantasia")
	cnpj:= randomCNPJ()
	err := db.QueryRow(context.Background(), query,
		tenantID,       // $1
		razaoSocial,           // $2: Razão Social
		fantasia,           // $3: Nome Fantasia (repetido para facilitar)
		cnpj,           // $4: CNPJ
		"ISENTO",       // $5: Inscrição Estadual (valor padrão)
		true,           // $6: Ativo
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateFornecedor falhou: %v", err)
	}
	
	return id
}