---
name: tenant-isolation-check
description: Audita queries SQL do projeto (database/queries/*.sql) em busca de falta de filtro por tenant_id, a principal fonte de vazamento de dados entre empresas neste SaaS multi-tenant. Use ao criar/alterar uma query, ao revisar um PR que mexe em database/queries ou database/repository, ou quando o usuário pedir uma auditoria de isolamento multi-tenant.
tools: Read, Grep, Bash
---

# Tenant Isolation Check

Este projeto é um SaaS multi-tenant onde quase toda tabela tem uma coluna
`tenant_id`. **Não existe Row-Level Security no Postgres** — o isolamento
entre empresas é feito 100% na aplicação, repetindo `tenant_id = $N` no
`WHERE` de cada query. Uma query nova sem esse filtro vaza dados de uma
empresa para outra.

Esta skill audita `database/queries/*.sql` (e, se pedido, um diff específico)
procurando queries que tocam tabelas com `tenant_id` mas não filtram por ele.

## Como funciona

### 1. Descubra quais tabelas têm `tenant_id`

```bash
grep -l "tenant_id" database/migrate/*.up.sql
```

Praticamente todas as tabelas de domínio têm. As exceções conhecidas (não
precisam de tenant_id porque são globais/catálogo) são:

- `planos` — catálogo de planos, compartilhado entre todos os tenants
- `schema_migrations` — controle interno do golang-migrate

### 2. Para cada arquivo em `database/queries/*.sql` (ou só os alterados, se for revisão de PR)

Cada bloco começa com `-- name: NomeDaQuery :one|:many|:exec|:execrows`.
Para cada bloco:

1. Identifique a(s) tabela(s) no `FROM`/`UPDATE`/`INSERT INTO`/`DELETE FROM`.
2. Se a tabela tem `tenant_id` (do passo 1), confira se o `WHERE` (ou o
   `VALUES`, no caso de INSERT) referencia `tenant_id`.
3. Se não referenciar, é um candidato a problema — **mas confira as exceções
   abaixo antes de reportar como bug**.

### 3. Exceções legítimas (NÃO reportar)

- **Rotas de super_admin** (`/api/painel/*` e `/api/master/*` em
  `internal/routers/rotasHttp.go`, protegidas por
  `middleware.VerificaSuperAdmin()`): são intencionalmente cross-tenant,
  porque o super admin da Radaptech gerencia TODAS as empresas. Ex:
  `ResumoDashboard`, `EmpresasRecentes`, `DadosEmpresas` (Empresa.sql),
  `MostrarUsuariosPainel`, `EditarUsuario`, `EditarStatusUsuario`
  (Usuario.sql), tudo em `planos.sql`.
- **`GetTenantBySubdomain`** (usada pelo `TenantMiddleware`): não pode
  filtrar por `tenant_id` porque o objetivo dela é justamente *descobrir*
  o `tenant_id` a partir do subdomínio.
- Queries que só tocam `planos` ou `schema_migrations`.

### 4. Reporte os achados

Para cada query suspeita, informe:
- Nome da query e arquivo (`database/queries/X.sql`)
- Por que ela parece precisar de `tenant_id` (qual tabela, e que essa tabela
  tem a coluna)
- Sugestão de correção (adicionar `AND tenant_id = sqlc.arg('tenant_id')` ou
  equivalente, seguindo o estilo já usado nas outras queries do mesmo
  arquivo)

Se nada for encontrado, diga isso explicitamente — não invente problema.

## Ao revisar uma query NOVA que está sendo escrita agora

Não espere o arquivo estar pronto: ao gerar ou editar uma query em
`database/queries/*.sql` como parte de outra tarefa (ex: "adicionar
endpoint"), rode essa checagen nela antes de seguir para
`sqlc generate` — é mais barato pegar aqui do que depois que o
repository/service já estão escritos em cima da query errada.
