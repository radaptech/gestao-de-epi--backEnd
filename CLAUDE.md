# CLAUDE.md

Back-end do **SGEPI** (Sistema de Gestão de EPIs) — API REST em Go, multi-tenant (SaaS), para
controle de estoque, entrega e devolução de Equipamentos de Proteção Individual.

Módulo Go: `github.com/davi-fernandesx/sistema-de-gestao-de-epi` · Go 1.26 · Gin · pgx/v5 · sqlc · PostgreSQL.

> O código, comentários, nomes de variáveis e mensagens de erro estão em **português**. Mantenha esse
> padrão ao escrever código novo.

---

## Comandos

```bash
# Build / run local (precisa de Postgres acessível e .env preenchido)
go build -o main .
go run .

# Testes (unitários + integração; -p 1 é obrigatório, testcontainers sobe Docker)
go test -v ./... -p 1

# Apenas um pacote
go test -v ./internal/helper/...
go test -v ./auth/...

# Gerar código do sqlc a partir de database/queries/*.sql + database/migrate/*.up.sql
sqlc generate

# Criar nova migração (gera par .up.sql / .down.sql numerado)
make migration nome-da-migracao

# Gerar docs Swagger (anotações nos controllers + main.go)
swag init
```

### Subir o ambiente completo (Docker)

O `docker-compose.dev.yml` vive **um nível acima**, em `sgeepi-infra/`, junto do front-end:

```bash
cd /home/davi/sge/sgeepi-infra
docker compose -f docker-compose.dev.yml --env-file gestao-de-epi--backEnd/.env up -d --build
# ou: make init
```

Sobe Traefik (roteia por subdomínio `*.localhost`), Postgres 16, pgAdmin (`:5050`),
Mailpit (`:8025`), a API (`Dockerfile.dev`, hot-reload via CompileDaemon) e o front React.

A API sozinha escuta em `:8080`. Via Traefik: `http://api.localhost` ou qualquer host com `/api`.

---

## Arquitetura

Fluxo em camadas, com injeção de dependência manual num container:

```
main.go
  └─ configs.Init          → conecta no Postgres (pgxpool) a partir do .env
  └─ RunMigrationPostgress → aplica golang-migrate no boot
  └─ SeedEmpresaMatriz     → cria plano/empresa/super_admin iniciais se não existirem
  └─ registra validator custom "cnpj"
  └─ routers.NewContainer(db) → monta repos → services → controllers
  └─ routers.ConfigurarRotas(router, container, db)

HTTP → middleware → controller/ → internal/service/ → database/repository/ → Postgres
```

| Camada | Pasta | Responsabilidade |
|---|---|---|
| Rotas + DI | `internal/routers/rotasHttp.go` | `Container` com todos os controllers; `ConfigurarRotas` define os grupos e middlewares |
| Controller | `controller/` | Bind/validação do JSON, extrai `tenantId`/`userId` do contexto, mapeia erro de domínio → status HTTP, anotações Swagger |
| Service | `internal/service/` | Regra de negócio, transações (`db.Begin` + `queries.WithTx`), montagem de DTOs, paginação |
| Repository | `database/repository/` | Wrapper fino sobre o sqlc; traduz erro do Postgres via `helper.TraduzErroPostgres` |
| SQL gerado | `database/repository/*.sql.go` | **Gerado pelo sqlc — nunca editar à mão** |
| Queries | `database/queries/*.sql` | Fonte de verdade das consultas (`-- name: X :one/:many/:exec/:execrows`) |
| Migrações | `database/migrate/` | golang-migrate, numeradas `NNNNNN_nome.up.sql` / `.down.sql` |
| Models/DTOs | `internal/model/` | Structs de entrada (`...Inserir`, binding tags) e saída (`...Dto`) |
| Helpers | `internal/helper/` | Erros de domínio, PDF (maroto), upload Supabase, tokens de auditoria, validador de CNPJ |
| Auth | `auth/` | Hash argon2id + geração de JWT |
| Config | `configs/` | Conexão, variáveis de ambiente, tipo `DataBr` |

### Distinção importante: `*_repository.go` vs `*.sql.go`

Em `database/repository/` convivem dois tipos de arquivo:

- `Epi.sql.go`, `Departamento.sql.go`, … → **gerados pelo sqlc**, sobrescritos a cada `sqlc generate`.
- `Epi_repository.go`, `Departamento_repository.go`, … → **escritos à mão**. Encapsulam o `*Queries`,
  expõem os métodos que o service consome e traduzem erros. É aqui que você mexe.

O construtor típico é `NewXRepository(pool *pgxpool.Pool)` guardando `q: New(pool)` e `db: pool`.

---

## Multi-tenancy — a regra mais importante do projeto

Cada empresa cliente é um **tenant** (`tabela empresas`). Praticamente toda tabela tem
`tenant_id INT NOT NULL REFERENCES empresas(id)`.

Como o tenant chega até a query:

1. O front envia o header **`X-Tenant-ID`** com o **subdomínio** da empresa (ex.: `acme`), não o ID.
2. `middleware.TenantMiddleware` resolve o subdomínio via `GetTenantBySubdomain` e faz
   `c.Set("tenantId", empresa.ID)`. Rejeita `www` e `api` como reservados.
3. O controller lê com `middleware.GetTenantID(c) (int32, bool)`.
4. O service repassa `tenantId` como parâmetro de **toda** query.

**Ao escrever qualquer query nova em `database/queries/`, o `WHERE` precisa incluir
`tenant_id = $N`.** Não existe RLS no banco — o isolamento é feito 100% na aplicação. Um `WHERE`
sem tenant vaza dados entre empresas.

O `userId` logado sai do JWT e é lido com `middleware.GetUserID(c)` (usado para auditoria:
`id_usuario_criacao`, `id_usuario_cancelamento`, etc.).

---

## Autenticação e autorização

- **Hash de senha**: argon2id (`auth.HashPassword` / `auth.HashCompare`).
- **JWT** (`auth.GerarJWT`): HS256, `JWT_SECRET` do ambiente, expira em 24 h. Claims:
  `sub` (id do usuário), `nome`, `role`, `tenantId`, `exp`, `iat`.
- **Transporte do token**: `middleware.AutenticacaoJWT` procura primeiro o **cookie `token`**
  (HttpOnly, `Secure`, `SameSite=None`, setado no login) e só depois cai para o header
  `Authorization: Bearer <token>`.
- O middleware popula no contexto: `userId`, `user_role`, `user_TenantId`, `user_nome`.

### Roles

| Role | Acesso |
|---|---|
| `colaborador` (default) | Rotas de leitura sob `/api` + cadastro de entrega/devolução |
| `admin` | Tudo acima + grupo `/api/gerencial` (`middleware.VerificaRole("admin")`) |
| `super_admin` | Grupos `/api/painel` e `/api/master` (`middleware.VerificaSuperAdmin`) |

### Os três grupos de rota

```go
/api/painel   → AutenticacaoJWT + VerificaSuperAdmin       // cadastro de empresas
/api/master   → AutenticacaoJWT + VerificaSuperAdmin       // dashboard global, planos, usuários
/api          → TenantMiddleware                            // login, logout, recuperação de senha
/api          → + AutenticacaoJWT + LoggerComUsuario        // operação do dia a dia
/api/gerencial→ + VerificaRole("admin")                     // escrita/cadastros
```

Atenção ao `gin.RouterGroup`: `api.Use(...)` é aplicado em cascata — as rotas registradas **depois**
de um `Use` herdam o middleware. `/api/painel` e `/api/master` são grupos separados de `r`, então
**não passam pelo `TenantMiddleware`** (super admin é global, não pertence a um tenant).

Rotas públicas: `GET /`, `GET /api` (health) e `GET /swagger/*any`.

---

## Domínio e regras de negócio

### Modelo de dados (principais tabelas)

```
empresas (tenant) ──┬── usuarios (role, ultimo_acesso, token_recuperacao_senha)
                    ├── departamento ── funcao ── funcionario (matricula, cpf)
                    ├── tipo_protecao ── epi ── tamanhos_epis ── tamanho
                    ├── fornecedor
                    ├── entrada_nf ── entrada_epi_item   (o "lote": quantidade_atual, validade)
                    ├── entrega_epi ── epis_entregues    (aponta para o lote de origem)
                    ├── motivo_devolucao (eh_descarte)
                    ├── devolucao (opcionalmente com troca: epi/tamanho/quantidade novos)
                    └── planos (limite_funcionarios / limite_usuarios / limite_epis, mensalidade)
```

### Estoque: controle por lote, FEFO

O saldo **não** é um contador único — vive em `entrada_epi_item.quantidade_atual`, por lote.

- **Entrada** (`EntradaService.Adicionar`): cria `entrada_nf` (cabeçalho) + N `entrada_epi_item`
  numa transação. `quantidade_atual` começa igual a `quantidade`. Valida data de emissão não futura
  e `data_validade >= data_fabricacao`.
- **Entrega** (`EntregaService.RegistrarEntrega`): para cada item pedido, `ListarLotesParaConsumo`
  seleciona lotes com `quantidade_atual > 0 AND data_validade >= CURRENT_DATE`, ordenados por
  `data_validade ASC` (**FEFO**) e com `FOR UPDATE` (trava a linha dentro da transação). Consome
  lote a lote via `AbaterEstoqueLote`, gerando uma linha em `epis_entregues` por lote consumido —
  ou seja, **um item pedido pode virar várias linhas entregues**. Se o saldo não fechar, retorna
  erro contendo `"estoque insuficiente"` e a transação sofre rollback.
- **Cancelamento de entrega**: marca a entrega e os itens como cancelados e **repõe** cada
  quantidade no lote exato de origem (`ReporEstoqueLote`).
- **Devolução** (`DevolucaoService.SalvarDevolucao`): valida o saldo em posse do funcionário
  (`ConsultarSaldoEpiFuncionario` = entregue − devolvido). Se o motivo tiver `eh_descarte = false`,
  repõe nos lotes respeitando o espaço disponível (`quantidade - quantidade_atual`); se for descarte,
  o item não volta ao estoque. Com `houve_troca = true`, dispara uma nova entrega na mesma transação.

### Assinatura digital e token de auditoria

Entregas e devoluções exigem assinatura. O fluxo no controller é sempre:

1. Gerar token de auditoria — `helper.GerarTokenAuditoria` (`ENT-…`) ou `GerarTokenDevolucao`
   (`DEVO-…`): SHA-256 de `NOME|FUNÇÃO|DEPTO|YYYY-MM-DD`, truncado em 16 caracteres.
2. `helper.UploadAssinaturaSupabase(base64, token, pasta)` — decodifica o base64, sobe como PNG no
   bucket Supabase e devolve a **URL pública**.
3. Substituir `input.Assinatura_Digital` pela URL antes de chamar o service. **O banco guarda a URL,
   nunca o base64.**

### Geração de PDF

`internal/helper/pdfHelperEntrega.go` e `pdfHelperDevolucao.go` usam **maroto/v2** para montar a
ficha de EPI (com código de barras e a imagem da assinatura baixada da URL, rotacionada quando
necessário). Rotas: `GET /api/gerencial/:matricula/ficha-pdf/:id` e
`GET /api/gerencial/devolucoes/:id/pdf`.

### Matrícula do funcionário

Gerada por **trigger no Postgres** (`trigger_gerar_matricula`, migração 000016): antes do INSERT,
se `matricula IS NULL`, calcula `MAX(matricula)+1` dentro do mesmo `tenant_id`. Não gere matrícula
na aplicação.

### Limites de plano

`UsuarioService.Registrar` compara `TotalDeUsuario` com `planos.limite_usuarios` e retorna
`helper.ErrLimiteExcedido` (→ HTTP 403). Regra análoga existe para funcionários.

### Importação via XLSX

Departamentos, funções e fornecedores aceitam upload de planilha (`multipart/form-data`, campo
`file`) usando **excelize**. A leitura acontece **no controller** (`ImportDepartamentoXLSX`,
`ImportarFuncaoXLSX`, `ImportFornecedor`): abre a primeira aba, pula tudo até encontrar a linha de
cabeçalho esperada (ex.: `"Nome do Departamento"`) e insere linha a linha — sem transação única, um
erro no meio aborta com 409/500 e deixa os anteriores gravados.

### Soft delete

Quase nada é apagado de verdade. O padrão é `ativo = FALSE` + `deletado_em`/`cancelada_em`
preenchidos, e os `SELECT` filtram por `ativo = TRUE` / `deletado_em IS NULL`. Vários filtros de
listagem expõem um booleano `cancelados`/`canceladas` para consultar o oposto.

---

## Convenções de código

### Tratamento de erros

`internal/helper/Errors.go` define os erros de domínio (`ErrNaoEncontrado`, `ErrDadoDuplicado`,
`ErrConflitoIntegridade`, `ErrEstoqueInsuficiente`, `ErrDataMenor`, `ErrLimiteExcedido`, …) e
`TraduzErroPostgres` converte `pgconn.PgError` por código (`23505` → `ErrDadoDuplicado`,
`23503` → `ErrConflitoIntegridade`, `23502` → `ErrCampoObrigatorio`, …).

Regra de ouro: **repository traduz, service propaga com `errors.Is`, controller decide o status**:

```go
if errors.Is(err, helper.ErrDadoDuplicado) {
    ctx.JSON(http.StatusConflict, gin.H{"error": "...", "detalhes": err.Error()})
    return
}
```

Formato de resposta de erro em uso: `{"error": "<mensagem amigável>", "detalhes": "<err.Error()>"}`.
(Existe `helper.HTTPError` para o Swagger, mas os handlers respondem com `gin.H`.)

### Controllers

Cada controller declara **sua própria interface de service** no topo do arquivo (interface do lado
do consumidor) e recebe a implementação por construtor. Os handlers são *closures*:
`func (c *XController) Acao() gin.HandlerFunc`.

### Datas

`configs.DataBr` é um `time.Time` com `UnmarshalJSON`/`MarshalJSON` no formato **`DD/MM/YYYY`**.
Toda data que entra ou sai da API usa esse tipo. Helpers: `configs.NewDataBrPtr(t)`, `.Time()`,
`.IsZero()`.

### Tipos nulos e decimais

- Colunas anuláveis viram `pgtype.Text` / `pgtype.Int4` / `pgtype.Date` / `pgtype.Timestamp`.
  Padrão de preenchimento: `pgtype.Text{String: v, Valid: v != ""}`.
- `numeric` do Postgres é mapeado para `*decimal.Decimal` (shopspring) via override no `sqlc.yaml`.
- DTOs de update parcial (PATCH) usam ponteiros (`*string`, `*int32`) + `COALESCE(sqlc.narg(...))`
  na query, para distinguir "não enviado" de "enviado vazio".

### Transações

Sempre no **service** (exceto `EntradaRepository.AdicionarCompleta`, que é a exceção):

```go
tx, err := s.db.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx)      // no-op depois do Commit
qtx := s.queries.WithTx(tx)
// ... usa qtx ...
return tx.Commit(ctx)
```

Métodos de repository que participam de transação recebem `qtx *repository.Queries` como parâmetro;
os de leitura simples usam o pool direto.

### Paginação

`service.Paginacao(FiltroPaginacao)` normaliza `pagina`/`quantidade` (limite máximo 100) e devolve
`Limit`/`Offset`/`PaginaAtual`. As queries trazem `COUNT(*) OVER() AS total_geral`, e o service lê
`rows[0].TotalGeral` para calcular `PaginaFinal = ceil(total/limit)`.

Os filtros vêm por query string com tags `form:"..."` e `ShouldBindQuery`.

### Evitando N+1

Padrão adotado em listagens com relacionamento 1-N (entregas→itens, EPIs→tamanhos): busca a lista
principal paginada, busca **todos** os filhos do tenant numa segunda query, monta um
`map[int32][]Filho` e faz o merge em memória. Está comentado como "possível gargalo futuro" —
mantenha o padrão, mas fique atento ao volume.

### Swagger

Anotações `// @Summary`, `// @Router`, `// @Security BearerAuth` ficam acima dos handlers;
os metadados globais estão em `main.go`. Rode `swag init` após alterar rotas — `docs/docs.go`,
`swagger.json` e `swagger.yaml` são gerados.

---

## Testes

- **Unitários**: `auth/hash_test.go` (argon2id) e `internal/helper/CNPJ_test.go` (table-driven com
  testify). Rápidos, sem dependências externas.
- **Integração**: `internal/service/testeIntegracaoEntregaEpi_test.go` sobe um
  **Postgres 15-alpine via testcontainers-go**, cria o schema com `criarTabelasPostgres`
  (`SetupTabelasTesteIntegracao.go`) e popula fixtures com os helpers `CreateEmpresa`,
  `CreateFuncionario`, `CreateEpi`, `CreateEntradaEpi`, … (`SetupDadosTesteIntegracao.go`).
  Cobrem entrega concorrente e consistência de estoque.
- **Requer Docker rodando.** Use sempre `-p 1` (containers não toleram pacotes em paralelo).
- Os arquivos de setup **não** têm build tag: `go test ./...` já os compila. O `oque-fazer.txt`
  menciona `-tags=integration`, mas essa tag não existe mais no código.
- CI (`.github/workflows/cicd.yml`): roda `go test -v ./... -p 1` nos pushes/PRs para
  `main`, `dev` e `homologacao`. **A action fixa Go 1.25.1 enquanto o `go.mod` pede 1.26** — os
  testes de integração falham nesse setup.

---

## Variáveis de ambiente

Copie `.env-example` para `.env` (o `.env` está no `.gitignore`).

| Variável | Uso |
|---|---|
| `DB_SERVER`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DATABASE` | Conexão pgxpool |
| `DB_SSLMODE` | Default `disable` se ausente |
| `JWT_SECRET` | **Obrigatória** — `AutenticacaoJWT` dá `panic` no boot se estiver vazia |
| `JWT_EXPIRATION` | Presente no `.env`, mas **não é lida**: a expiração está fixa em 24 h no código |
| `SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY`, `SUPABASE_BUCKET` | Upload das assinaturas |
| `RESEND_API_KEY` | Envio de e-mail de recuperação de senha (Resend) |
| `URL_FRONTEND_FORMATO` | Template do link de redefinição, ex.: `http://%s:3000` (`%s` = subdomínio) |
| `SUPER_ADMIN_EMAIL`, `SUPER_ADMIN_PASSWORD` | Seed do super admin no boot |

---

## Fluxos comuns de manutenção

### Adicionar um endpoint

1. `database/queries/X.sql` — nova query com `-- name:` e `WHERE tenant_id = $N`.
2. `sqlc generate`.
3. Método no `database/repository/X_repository.go` (traduzindo o erro).
4. Adicione o método à interface do repository declarada no service e implemente a regra.
5. Adicione o método à interface de service declarada no controller e escreva o handler + anotações.
6. Registre a rota no grupo correto em `internal/routers/rotasHttp.go`.
7. `swag init`.

### Adicionar uma coluna/tabela

1. `make migration nome-descritivo` → edite o par `.up.sql` / `.down.sql`.
2. `sqlc generate` (o sqlc lê o schema a partir das próprias migrações `.up.sql`).
3. As migrações rodam automaticamente no boot do `main.go` (`RunMigrationPostgress`).

### Registrar um validador custom

`main.go` já registra a tag `cnpj` (`helper.ValidateCNPJ`) no engine do validator do Gin. Novos
validadores entram no mesmo bloco.

---

## Armadilhas conhecidas

- **O binário compilado `main` (~58 MB) já esteve versionado no Git.** Hoje está no `.gitignore` e
  removido do índice (`git rm --cached main`) — o arquivo continua no disco. Não o adicione de volta.
- **Migração 000019 não tem `.down.sql`**, então `migrate down` quebra ao chegar nela.
- **A migração 000001 usa o nome `000001_CreateTables.sql.up.sql`** (com `.sql` extra no meio) —
  siga o padrão limpo `NNNNNN_nome.up.sql` nas novas.
- **Os targets `migrate-up`/`migrate-down` do `makefile` chamam `go run main.go Up|Down`**, mas o
  `main.go` não lê argumentos de linha de comando — rodar isso sobe a aplicação inteira. Migrações
  são aplicadas no boot; para reverter, use a CLI do `golang-migrate` diretamente.
- **`golang-migrate` procura `file://database/migrate` relativo ao working dir.** Rodar o binário de
  outra pasta quebra o boot (por isso o `Dockerfile` copia `database/migrate` para junto dele).
- **`repository.Queries` não tem `sqlc.embed` nem RLS**: repetir `tenant_id` em cada query é
  responsabilidade sua.
- **Muito `fmt.Printf` de debug com emoji** nos services de Entrega/Devolução/EPI. É o estilo atual do
  projeto; se for limpar, faça em commit separado.
- **CORS** (`middleware/cors.go`) libera `localhost`, `*.localhost` e `*.radaptech.com.br`, com
  `AllowCredentials: true` (necessário para o cookie HttpOnly). Um domínio novo precisa entrar no
  `AllowOriginFunc`.
- **`SecurityHeaders` aplica `Content-Security-Policy: default-src 'self'`** em todas as respostas,
  inclusive na UI do Swagger.

---

## Git

Branch principal: `main`. Branch de trabalho: `dev`. Existe também `homologacao` e branches de
feature (`feature/HttpOnlyCookie`, `feature/importacaoCSV/XLSX`). Mensagens de commit em português,
no imperativo ("adiciona suporte a…", "atualiza dependências e…").
