# Arquitetura — SGEPI Backend

API REST em Go para o **SGEPI** (Sistema de Gestão de EPIs), um SaaS multi-tenant onde cada
empresa cliente controla seu próprio estoque, entregas e devoluções de Equipamentos de Proteção
Individual, isolada das demais dentro do mesmo banco de dados.

Stack: Go 1.26 · [Gin](https://github.com/gin-gonic/gin) (HTTP) · [pgx/v5](https://github.com/jackc/pgx)
(driver Postgres) · [sqlc](https://sqlc.dev) (SQL → Go tipado) · PostgreSQL · [golang-migrate](https://github.com/golang-migrate/migrate)
(migrações) · [maroto/v2](https://github.com/johnfercher/maroto) (geração de PDF) · Supabase Storage
(upload de assinaturas).

---

## Visão em camadas

```
HTTP request
   │
   ▼
middleware global (CORS, SecurityHeaders)
   │
   ▼
middleware de rota (TenantMiddleware / AutenticacaoJWT / LoggerComUsuario / VerificaRole / LimitarPorIP)
   │
   ▼
controller/            → bind + valida JSON, lê tenantId/userId do contexto, chama o service,
                          mapeia erro de domínio → status HTTP, hospeda as anotações Swagger
   │
   ▼
internal/service/      → regra de negócio, transações (db.Begin + queries.WithTx),
                          montagem de DTOs de saída, paginação
   │
   ▼
database/repository/   → wrapper fino sobre o código gerado pelo sqlc;
                          traduz erro do Postgres (helper.TraduzErroPostgres)
   │
   ▼
PostgreSQL
```

Cada camada só conhece a camada imediatamente abaixo. Controllers nunca tocam em `pgxpool.Pool`
ou em queries geradas — sempre passam pelo service. Services nunca montam SQL na mão — sempre
passam pelo repository.

### Por que essa separação

- **Controller fino**: toda a lógica de "o que fazer" fica no service, então o mesmo service
  poderia, em teoria, ser reaproveitado por outro transporte (CLI, worker, gRPC) sem reescrever
  regra de negócio.
- **Repository traduz erros do Postgres em erros de domínio** (`helper.ErrDadoDuplicado`,
  `helper.ErrConflitoIntegridade`, …) logo na saída do banco. Dali para cima (service, controller)
  ninguém mais lida com `pgconn.PgError` — só com `errors.Is`.
- **sqlc gera código, não se edita à mão.** A fonte de verdade das queries é
  `database/queries/*.sql`; o `.sql.go` correspondente é build artifact.

---

## Boot da aplicação (`main.go`)

Antes de qualquer coisa, `main()` olha `os.Args`: com o argumento `backup-banco` ele executa
`ExecutarBackupBanco` (dump do banco → Cloudflare R2, ver [`BACKUP.md`](./BACKUP.md)) e encerra,
**sem subir a API**. Sem argumento, segue a sequência normal abaixo:

1. `router.SetTrustedProxies([]string{"172.16.0.0/12"})` — só confia em `X-Forwarded-For` vindo da
   rede interna do proxy reverso (Traefik). Sem isso, o rate limit por IP (`middleware.LimitarPorIP`)
   poderia ser burlado por um cliente forjando o próprio IP no header.
2. `middleware.CorsConfig()` e `middleware.SecurityHeaders()` são registrados **globalmente** no
   engine do Gin (`router.Use(...)`), antes de qualquer rota — valem até para `/swagger/*any`.
3. `configs.Init.InitAplicattion()` — lê `.env` (`godotenv`) e abre o pool `pgxpool.Pool` contra o
   Postgres.
4. `ConexaoDbPostgres.RunMigrationPostgress(db)` — aplica todas as migrações pendentes de
   `database/migrate` via `golang-migrate`, usando uma conexão emprestada do pool
   (`stdlib.OpenDBFromPool`). Roda **toda vez que a API sobe**, inclusive em produção — não existe
   passo manual de deploy para migração.
5. `SeedEmpresaMatriz(db)` (`seed.go`) — garante que existam o plano, a empresa "matriz" e o
   usuário `super_admin` iniciais, criando-os apenas se ainda não existirem. Idempotente.
6. Registro do validator custom `cnpj` (`helper.ValidateCNPJ`) no engine do `go-playground/validator`
   usado pelo Gin — depois disso, qualquer DTO pode usar a tag `binding:"cnpj"`.
7. `routers.NewContainer(db)` — monta manualmente toda a árvore de dependências (repository → service
   → controller) para cada domínio e devolve um `Container` com todos os controllers prontos.
8. `routers.ConfigurarRotas(router, container, db)` — registra as rotas e aplica os middlewares por
   grupo.
9. `router.Run(":8080")`.

Não há framework de injeção de dependência — o `Container` em `internal/routers/rotasHttp.go` é a
"raiz de composição": cada domínio segue o padrão `repoX := repository.NewXRepository(db)` →
`serviceX := service.NewXService(repoX, db)` → `controller.NewXController(serviceX)`.

---

## Multi-tenancy

Ver `CLAUDE.md` na raiz do projeto para a regra completa — resumo:

- Toda tabela de negócio tem `tenant_id INT NOT NULL REFERENCES empresas(id)`.
- O front manda o subdomínio da empresa no header `X-Tenant-ID` (não o ID numérico).
- `middleware.TenantMiddleware` resolve o subdomínio → `empresa.ID` e guarda em `c.Set("tenantId", ...)`.
- **Não existe Row-Level Security no Postgres.** O isolamento entre empresas é 100% responsabilidade
  da aplicação: toda query nova em `database/queries/*.sql` precisa filtrar por `tenant_id = $N`.
- `super_admin` (rotas `/api/painel` e `/api/master`) é global e não passa pelo `TenantMiddleware` —
  não pertence a nenhum tenant específico.

Detalhes de cada middleware envolvido: [`MIDDLEWARES.md`](./MIDDLEWARES.md).

---

## Modelo de dados e regras de estoque

Ver [`MODELO_DE_DADOS.md`](./MODELO_DE_DADOS.md) para o schema completo (tabelas, colunas, FKs,
triggers) e [`FLUXOS_DE_NEGOCIO.md`](./FLUXOS_DE_NEGOCIO.md) para a lógica de entrada/entrega/
devolução de estoque (controle por lote, FEFO, assinatura digital, etc).

---

## Tratamento de erros

`internal/helper/Errors.go` define os erros de domínio e `TraduzErroPostgres` converte
`pgconn.PgError` por código SQLSTATE:

| Código Postgres | Erro de domínio |
|---|---|
| `23505` (unique violation) | `ErrDadoDuplicado` |
| `23503` (foreign key violation) | `ErrConflitoIntegridade` |
| `23502` (not null violation) | `ErrCampoObrigatorio` |
| `23514` (check violation) | erro genérico "regra de validação do banco violada" |
| `40P01` (deadlock) | erro genérico "sistema ocupado, tente novamente" |

Fluxo: **repository traduz → service propaga com `errors.Is` → controller decide o status HTTP**.
Resposta de erro padrão: `{"error": "<mensagem amigável>", "detalhes": "<err.Error()>"}`.

---

## Documentação Swagger

Gerada por `swag init` a partir das anotações `// @Summary`, `// @Router`, `// @Security BearerAuth`
nos handlers de `controller/*.go` e dos metadados globais em `main.go`. Saída em `docs/docs.go`,
`docs/swagger.json`, `docs/swagger.yaml` — **não editar esses três arquivos à mão**, são gerados.
Servida em `/swagger/*any` (rota pública, sem autenticação).

---

## Ambientes e execução

- **Local sem Docker**: precisa de Postgres acessível e `.env` preenchido (copiar de `.env-example`).
  `go run .` ou `go build -o main . && ./main`.
- **Docker completo**: `docker-compose.dev.yml` vive um nível acima, em `sgeepi-infra/`, junto do
  front-end. Sobe Traefik (roteamento por subdomínio `*.localhost`), Postgres 16, pgAdmin, Mailpit
  e a API com hot-reload (`Dockerfile.dev` + CompileDaemon). A API sozinha escuta em `:8080`; via
  Traefik, `http://api.localhost`.
- **CI** (`.github/workflows/cicd.yml`): roda `go test -v ./... -p 1` em push/PR para `main`, `dev`
  e `homologacao`.

## Variáveis de ambiente

| Variável | Uso |
|---|---|
| `DB_SERVER`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DATABASE` | Conexão `pgxpool` |
| `DB_SSLMODE` | Default `disable` se ausente |
| `JWT_SECRET` | **Obrigatória** — a API dá `panic` no boot se estiver vazia |
| `JWT_EXPIRATION` | Presente no `.env`, mas não é lida; expiração fixa em 24h no código |
| `SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY`, `SUPABASE_BUCKET` | Upload das imagens de assinatura |
| `RESEND_API_KEY` | Envio de e-mail de recuperação de senha |
| `URL_FRONTEND_FORMATO` | Template do link de redefinição de senha, ex.: `http://%s:3000` |
| `SUPER_ADMIN_EMAIL`, `SUPER_ADMIN_PASSWORD` | Seed do super admin no boot |
| `R2_IDCLOUDFLARE`, `R2_KEYID`, `R2_SECRETKEY`, `R2_BUCKET_NAME_BACKUPS` | Backup do banco no Cloudflare R2 — só usadas pelo subcomando `./main backup-banco` ([`BACKUP.md`](./BACKUP.md)) |

## Armadilhas conhecidas

- O binário compilado `main` (~58 MB) já esteve versionado no Git; está no `.gitignore` hoje, mas
  não o adicione de volta.
- Migração `000019` não tem `.down.sql` — `migrate down` quebra ao chegar nela.
- `golang-migrate` procura `file://database/migrate` **relativo ao working dir** do processo — rodar
  o binário de outra pasta quebra o boot.
- `repository.Queries` não tem RLS nem `sqlc.embed`: repetir `tenant_id` em cada query nova é
  responsabilidade de quem escreve a query.
- Os targets `migrate-up`/`migrate-down` do `makefile` chamam `go run main.go Up|Down`, mas o único
  argumento que `main.go` reconhece é `backup-banco` — qualquer outro sobe a aplicação inteira, não
  reverte nada. Para reverter, use a CLI do `golang-migrate` direto.
- O `Dockerfile` do builder precisa acompanhar a versão do `go.mod`: com `go 1.26.5` no `go.mod` e
  `golang:1.26.1-alpine` no `FROM`, o `go mod download` quebra o build. Vale para o
  `Dockerfile.dev` também.
