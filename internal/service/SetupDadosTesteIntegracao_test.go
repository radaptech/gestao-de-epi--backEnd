package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// --- UTILITÁRIOS ---

func randomString(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func randomInt() int32 {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return int32(r.Intn(9999))
}

// Helper para gerar um "CNPJ" numérico aleatório de 14 dígitos
func randomCNPJ() string {

	rand.NewSource(time.Now().UnixNano())
	return fmt.Sprintf("%014d", rand.Int63n(99999999999999))
}

// --- 1. EMPRESA (TENANT) ---

func CreatePlanos(t *testing.T, db *pgxpool.Pool) int64 {

	var id int64
	query := `INSERT INTO planos (nome, mensalidade, limite_funcionarios, limite_usuarios, limite_epis, status, descricao)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id;`

	nome := "plano teste"
	mensalidade := 1000
	limite_funcionarios := 1000
	limite_usuarios := 1000
	limite_epis := 1000
	status := "Ativo"
	descricao := "teste descricao"

	err := db.QueryRow(context.Background(), query,
		nome, mensalidade, limite_funcionarios, limite_usuarios, limite_epis, status, descricao).Scan(&id)
	if err != nil {
		t.Fatalf("erro ao criar plano de teste: %v", err)

	}

	return id
}

func CreateEmpresa(t *testing.T, db *pgxpool.Pool, idplanos int64) int64 {
	var id int64
	query := `
		INSERT INTO empresas (
			nome_fantasia,
			razao_social,
			cnpj,
			responsavel,
			email,
			telefone,
			plano_id,
			status,
			vencimento,
			observacoes,
			subdominio
		) VALUES (
		
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11
		)
		RETURNING id;
	`
	// Gera dados aleatórios para satisfazer as constraints UNIQUE
	nano := time.Now().UnixNano()
	nomeFantasia := fmt.Sprintf("Radap Client %d", nano)
	razaoSocial := fmt.Sprintf("Radap Tech Clientes Ltda %d", nano)
	cnpj := fmt.Sprintf("%d", nano) // CNPJ único (fake)
	responsavel := fmt.Sprintf("responsavel%d", nano)
	email := fmt.Sprintf("email%d", nano)
	telefone := "1234567899"
	planoisd := idplanos
	status := "Ativa"
	vencimentos := time.Now()
	obsv := "vghjdfhjbf"
	subdominio := fmt.Sprintf("cliente%d", nano) // Subdominio único

	err := db.QueryRow(context.Background(), query,
		nomeFantasia,
		razaoSocial,
		cnpj,
		responsavel,
		email,
		telefone,
		planoisd,
		status,
		vencimentos,
		obsv,
		subdominio,
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
		INSERT INTO usuarios (tenant_id, nome, email, senha_hash, role) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	nome := randomString("Usuario Teste")
	// O email deve ser único por tenant
	email := fmt.Sprintf("user_%d@radap.com", time.Now().UnixNano())
	senha := "hash_senha_segura"
	role := "admin"

	err := db.QueryRow(context.Background(), query,
		tenantID,
		nome,
		email,
		senha,
		role,
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
		INSERT INTO departamento (tenant_id, nome) 
		VALUES ($1, $2)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query,
		tenantID,
		randomString("Dep"),
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateDepartamento falhou: %v", err)
	}
	return id
}

func CreateFuncao(t *testing.T, db *pgxpool.Pool, idDep, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO funcao (tenant_id, nome, IdDepartamento) 
		VALUES ($1, $2, $3) 
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query,
		tenantID,
		randomString("Func"),
		idDep,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateFuncao falhou: %v", err)
	}
	return id
}

func CreateFuncionario(t *testing.T, db *pgxpool.Pool, IdDepartamento, IdFuncao, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO funcionario (tenant_id, nome, matricula,IdDepartamento, IdFuncao, cpf) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	nome := randomString("Funcionario")
	matriculo:= randomInt()
	// cpf precisa caber em VARCHAR(11) e ser único por tenant (UNIQUE(tenant_id, cpf))
	cpf := fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)

	err := db.QueryRow(context.Background(), query,
		tenantID,
		nome,
		matriculo,
		IdFuncao,
		IdDepartamento,
		cpf,
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
		INSERT INTO tipo_protecao (tenant_id, nome) 
		VALUES ($1, $2)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query,
		tenantID,
		randomString("Protecao"),
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateProtecao falhou: %v", err)
	}
	return id
}

func CreateTamanho(t *testing.T, db *pgxpool.Pool, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO tamanho (tenant_id, tamanho) 
		VALUES ($1, $2)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query,
		tenantID,
		randomString("Tam"),
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
			validade_CA, IdTipoProtecao, alerta_minimo
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`
	r := randomInt()
	nome := randomString("Luva X")
	fabricante := randomString("Fabr Y")
	ca := fmt.Sprintf("%d", r) // CA único por tenant
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
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEpi falhou: %v", err)
	}
	return id
}

// --- 5. ESTOQUE E MOVIMENTAÇÃO ---

func CreateEntradaNfEpi(t *testing.T, db *pgxpool.Pool, tenantID, iduser int64, idfornecedor int64) int64 {
	var id int64

	// Atenção: Adicionado id_usuario_criacao, nota_fiscal e tenant_id
	query := `
		INSERT INTO entrada_nf (tenant_id, nota_fiscal_numero, nota_fiscal_serie, data_emissao, id_usuario_criacao, Idfornecedor)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	nfAleatoria := randomInt()
	iduserCriacao := iduser

	// 2. Converta para string (VARCHAR)
	notaFiscal := strconv.Itoa(int(nfAleatoria))
	notaFiscalSerie := strconv.Itoa(int(1))
	// Constraint: UNIQUE (nota_fiscal_numero, nota_fiscal_serie, fornecedor)
	// Usamos randomString para garantir unicidade nos testes
	err := db.QueryRow(context.Background(), query,
		tenantID,
		notaFiscal,
		notaFiscalSerie,
		time.Now(), // Data Entrada (Tipo Date no banco, Driver converte)
		iduserCriacao,
		idfornecedor,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEntradaNf falhou: %v", err)
	}
	return id
}

func CreateEntradaEpi(t *testing.T, db *pgxpool.Pool, tenantid, idEpi, idTamanho, iduser, IDentradaNf int64) int64 {

	var id int64
	query := `
	
		INSERT INTO entrada_epi_item (
   		 tenant_id, entrada_nf_id, id_epi, id_tamanho, quantidade, 
   		 quantidade_atual, data_fabricacao, data_validade, lote, valor_unitario,id_usuario_criacao
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id;
	`

	quantidade, quantidade_atual := 100, 100
	lote := randomString("lo")
	valor_unitario := decimal.NewFromFloat(56.99)

	err := db.QueryRow(context.Background(), query,
		tenantid,
		IDentradaNf,
		idEpi,
		idTamanho,
		quantidade,
		quantidade_atual,
		time.Now().AddDate(-2, 0, 0),
		time.Now().AddDate(3, 0, 0),
		lote,
		valor_unitario,
		iduser,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEntradaEpi falhou: %v", err)
	}
	return id
}

func CreateEntradaEpi1(t *testing.T, db *pgxpool.Pool, tenantid, idEpi, idTamanho, iduser, IDentradaNf int64) int64 {

	var id int64
	query := `
	
		INSERT INTO entrada_epi_item (
   		 tenant_id, entrada_nf_id, id_epi, id_tamanho, quantidade, 
   		 quantidade_atual, data_fabricacao, data_validade, lote, valor_unitario,id_usuario_criacao
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id;
	`

	quantidade, quantidade_atual := 1, 1
	lote := randomString("lo")
	valor_unitario := decimal.NewFromFloat(56.99)

	err := db.QueryRow(context.Background(), query,
		tenantid,
		IDentradaNf,
		idEpi,
		idTamanho,
		quantidade,
		quantidade_atual,
		time.Now().AddDate(-2, 0, 0),
		time.Now().AddDate(3, 0, 0),
		lote,
		valor_unitario,
		iduser,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEntradaEpi falhou: %v", err)
	}
	return id
}

func CreateEntregaEpi(t *testing.T, db *pgxpool.Pool, idFuncionario, idUserEntrega, tenantID int64) int64 {
	var id int64
	// Adicionado token_validacao e id_usuario_entrega
	query := `
		INSERT INTO entrega_epi (
		tenant_id, 
		IdFuncionario, data_entrega, assinatura, IdTroca, token_validacao, id_usuario_entrega
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;
	`
	assinatura := randomString("AssinaturaBase64")
	token := randomString("TokenValidacao")
	idtroca:= 1

	err := db.QueryRow(context.Background(), query,
		tenantID,
		idFuncionario,
		time.Now(),
		assinatura,
		idtroca,
		token,
		idUserEntrega,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEntregaEpi falhou: %v", err)
	}
	return id
}

func CreateEpiEntregues(t *testing.T, db *pgxpool.Pool, IDEntregaCabecalho, idEntradaItem, IdEpi, IdTamanho, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO epis_entregues (
			tenant_id, id_entrega_cabecalho, id_entrada_item, id_epi, id_tamanho, quantidade
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`
	quantidade := 10

	err := db.QueryRow(context.Background(), query,
		tenantID,
		IDEntregaCabecalho,
		idEntradaItem,
		IdEpi,
		IdTamanho,
		quantidade,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateEpiEntregues falhou: %v", err)
	}
	return id
}

func CreateMotivoDevolucao(t *testing.T, db *pgxpool.Pool, motivo string, geraDescarte bool, tenantID int64) int64 {
	var id int64
	query := `
		INSERT INTO motivo_devolucao (tenant_id, motivo, gera_descarte)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	err := db.QueryRow(context.Background(), query,
		tenantID,
		motivo,
		geraDescarte,
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
			inscricao_estadual
		) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`
	razaoSocial := randomString("razao")
	fantasia := randomString("fantasia")
	cnpj := randomCNPJ()
	err := db.QueryRow(context.Background(), query,
		tenantID,    // $1
		razaoSocial, // $2: Razão Social
		fantasia,    // $3: Nome Fantasia (repetido para facilitar)
		cnpj,        // $4: CNPJ
		"ISENTO",    // $5: Inscrição Estadual (valor padrão     
	).Scan(&id)

	if err != nil {
		t.Fatalf("Helper CreateFornecedor falhou: %v", err)
	}

	return id
}
