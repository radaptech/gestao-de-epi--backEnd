# Arquitetura do Backend - Sistema de Gestão de EPIs (SaaS Multi-Tenant)

## Visão Geral
Este backend é uma API RESTful escrita em Go utilizando o framework Gin, projetada para funcionar como um serviço **Software-as-a-Service (SaaS)** com arquitetura **multi-tenant**. Cada cliente (empresa) possui seus dados isolados no mesmo banco de dados PostgreSQL, utilizando um identificador de tenant (`tenant_id`) para separação lógica.

## Arquitetura de Separação Multi-Tenant no Banco de Dados

### Tabela Mãe: `empresas`
- Armazena os dados cadastrais de cada cliente (tenant).
- Campos relevantes:
  - `id` (SERIAL PRIMARY KEY) – usado como `tenant_id` em todas as outras tabelas.
  - `subdominio` (VARCHAR UNIQUE) – utilizado pelo middleware para identificar o tenant via header `X-tenant-ID`.
  - `cnpj` (VARCHAR UNIQUE) – identificador fiscal único.
  - `nome_fantasia`, `razao_social`, `ativo`, `criado_em`.

### Demais Tabelas
Todas as tabelas de domínio possuem:
- Uma coluna `tenant_id` (INT NOT NULL) que referencia `empresas.id`.
- Um `FOREIGN KEY (tenant_id) REFERENCES empresas(id)`.
- Quando aplicável, restrições de unicidade **escopadas ao tenant** (ex: `UNIQUE (tenant_id, cnpj)` na tabela de empresas já é global, mas em tabelas como `funcionario` temos `UNIQUE (tenant_id, matricula)` e na tabela `epi` temos `UNIQUE (tenant_id, CA)`).
- Colunas de controle: `ativo` (BOOLEAN), `deletado_em` (TIMESTAMP NULL) para soft delete.

### Exemplos de Estrutura (migração inicial)
```sql
CREATE TABLE funcionarios (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    nome VARCHAR(100) NOT NULL,
    matricula VARCHAR(20) NOT NULL,
    IdFuncao INT NOT NULL,
    IdDepartamento INT NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    deletado_em TIMESTAMP NULL,
    FOREIGN KEY (tenant_id) REFERENCES empresas(id),
    FOREIGN KEY (IdFuncao) REFERENCES funcao(id),
    FOREIGN KEY (IdDepartamento) REFERENCES departamento(id),
    UNIQUE (tenant_id, matricula)
);
```

### Como o Tenant é Identificado (Middleware)
1. O cliente deve enviar o header `X-tenant-ID` contendo o **subdomínio** da empresa (ex: `acme`).
2. Middleware `TenantMiddleware` (em `middleware/TenantId.go`):
   - Lê o header.
   - Ignora valores reservados (`www`, `api`).
   - Consulta a tabela `empresas` pelo `subdominio`.
   - Em caso de sucesso, armazena o `empresa.id` no contexto do Gin bajo a chave `"tenantId"`.
   - Em caso de falha, retorna erro HTTP apropriado (400, 404, 403).
3. A função helper `GetTenantID(c *gin.Context)` permite aos controllers obter o `tenant_id` de forma tipada.
4. Todas as queries geradas pelo `sqlc` (ou repositórios) esperam o `tenant_id` como primeiro parâmetro, garantindo isolamento.

### Benefícios
- Isolamento de dados: uma empresa nunca vê os dados de outra.
- Escalabilidade horizontal: novo tenant é apenas um insert na tabela `empresas`.
- Facilidade de backup/restore por tenant (filtrando por `tenant_id`).
- Manutenção simplificada: schema único para todos os tenants.

## Principais Endpoints da API (Go)

A API está estruturada em grupos de rotas, conforme definido em `internal/routers/rotasHttp.go`:

### 1. Grupo Público (sem autenticação)
- `GET /` – mensagem de boas-vindas.
- `GET /api` – status da API.
- `GET /swagger/*any` – documentação Swagger UI.

### 2. Grupo que Requer Tenant ID (header `X-tenant-ID`)
- `POST /api/login` – autenticação de usuário (retorna JWT).
- `POST /api/cadastro` – registro de novo usuário.
- `POST /api/esqueci-minha-senha` – salva token de recuperação.
- `POST /api/redefinir-senha` – redefine senha usando token.

### 3. Grupo Protegido (JWT + Tenant)
Requer token JWT válido no header `Authorization: Bearer <token>`.

#### Acessíveis por colaborador e admin
- `GET /api/me` – perfil do usuário logado.
- `GET /api/departamentos` – lista departamentos do tenant.
- `GET /api/quantidade-epi` – quantidade total de EPIs em estoque.
- `GET /api/saldo-epi` – valor total em estoque.
- `GET /api/funcoes` – lista de funções.
- `GET /api/funcionarios` – lista de funcionários.
- `GET /api/funcionario/:matricula` – funcionário por matrícula.
- `GET /api/funcionarios-dashbord` – dados para dashboard.
- `GET /api/funcionario-estoque` – funcionário com histórico de entregas.
- `GET /api/tamanhos` – lista de tamanhos.
- `GET /api/tamanhos-id-epi/:id` – tamanhos vinculados a um EPI.
- `GET /api/tamanho/:id` – tamanho por ID.
- `GET /api/protecoes` – lista de tipos de proteção.
- `GET /api/protecao/:id` – proteção por ID.
- `GET /api/epis` – lista de EPIs.
- `GET /api/epi/:id` – EPI por ID.
- `GET /api/epis-dashbord` – EPIs para dashboard.
- `GET /api/funcionarios/:id/epis` – EPIs entregues a um funcionário.
- `GET /api/entradas` – lista de entradas de EPI.
- `GET /api/entradas-dashbord` – entradas para dashboard.
- `GET /api/entradas-estoque` – entradas com detalhes de estoque.
- `GET /api/fornecedores` – lista de fornecedores.
- `GET /api/entregas` – lista de entregas.
- `POST /api/cadastro-entregas` – nova entrega.
- `GET /api/entregas-dashbord` – entregas para dashboard.
- `GET /api/entrega-itens-dashbord` – itens das entregas.
- `POST /api/cadastro-devolucao` – nova devolução.
- `GET /api/devolucao` – lista de devoluções.
- `POST /api/cadastrar-motivo-devolucao` – cadastra motivo de devolução.
- `GET /api/motivos` – lista de motivos de devolução.

#### Apenas para usuários com role "admin" (grupo `/gerencial`)
- `DELETE /api/gerencial/departamento/:id`
- `PUT /api/gerencial/departamento/:id`
- `POST /api/gerencial/cadastro-departamento`
- `DELETE /api/gerencial/funcao/:id`
- `PUT /api/gerencial/funcao/:id`
- `POST /api/gerencial/cadastro-funcao`
- `DELETE /api/gerencial/funcionario/:id`
- `PATCH /api/gerencial/funcionario/:id`
- `POST /api/gerencial/cadastro-funcionario`
- `POST /api/gerencial/cadastro-tamanho`
- `DELETE /api/gerencial/tamanho/:id`
- `POST /api/gerencial/cadastro-protecao`
- `DELETE /api/gerencial/protecao/:id`
- `DELETE /api/gerencial/epi/:id`
- `PATCH /api/gerencial/epi/:id`
- `POST /api/gerencial/cadastro-epi`
- `POST /api/gerencial/cadastrar-entrada`
- `DELETE /api/gerencial/entrada/:id`
- `POST /api/gerencial/cadastro-fornecedores`
- `DELETE /api/gerencial/fornecedor/:id`
- `PATCH /api/gerencial/fornecedor/:id`
- `DELETE /api/gerencial/entrega/:id`
- `GET /api/gerencial/:matricula/ficha-pdf/:id` – gera PDF da ficha de EPI.
- `GET /api/gerencial/usuarios` – lista todos os usuários do tenant.
- `GET /api/gerencial/devolucoes/:id/pdf` – gera PDF da devolução.

### 4. Grupo Painel (Super Admin)
Requer JWT + role `super-admin` (verificado por `middleware.VerificaSuperAdmin()`).
- `POST /api/painel/cadastrar-planos` – cria plano de assinatura.
- `GET /api/painel/planos` – lista planos.
- `PUT /api/painel/planos/:id` – atualiza plano.
- `PATCH /api/painel/planos/:id/status` – altera status do plano.
- `POST /api/painel/cadastrar-empresa` – cadastra nova empresa (tenant).

## Segurança
- **Autenticação**: JWT (middleware `AutenticacaoJWT`).
- **Autorização**: roles verificadas por `middleware.VerificaRole`.
- **Isolamento de Tenant**: middleware `TenantMiddleware`.
- **Headers de Segurança**: `SecurityHeaders.go` (CSP, HSTS, etc.).
- **CORS**: configurável via `cors.go`.

## Tecnologias
- **Linguagem**: Go 1.22+
- **Framework Web**: Gin-Gonic
- **Banco de Dados**: PostgreSQL
- **ORM/Query Builder**: sqlc (consultas geradas a partir de SQL)
- **Validação**: go-playground/validator v10 (com validador customizado de CNPJ)
- **Documentação**: Swagger (via swaggo/gin-swagger)
- **Variáveis de Ambiente**: carregadas via `github.com/joho/godotenv` (implícito em `VariaveisDeAmbiente.go`)
- **Migrações**: golang-migrate/migrate

## Observações
- O campo `subdominio` na tabela `empresas` é usado como identificador amigável no header `X-tenant-ID`.
- Todas as mutations (INSERT, UPDATE, DELETE) devem incluir o `tenant_id` nas condições de WHERE para garantir isolamento.
- O sistema utiliza soft delete (`deletado_em`) em muitas tabelas para histórico.
- Controllers recebem serviços injetados (Dependency Injection via Container em `routers.NewContainer`).
