# Documentação — SGEPI Backend

Índice da documentação técnica do backend. Para visão geral do projeto, comandos e convenções de
código, veja o `CLAUDE.md` na raiz do repositório.

| Documento | Conteúdo |
|---|---|
| [`ARQUITETURA.md`](./ARQUITETURA.md) | Camadas da aplicação, boot (`main.go`), multi-tenancy, tratamento de erros, ambientes/variáveis, armadilhas conhecidas |
| [`MIDDLEWARES.md`](./MIDDLEWARES.md) | Cada middleware (`CORS`, `SecurityHeaders`, `TenantMiddleware`, `AutenticacaoJWT`, `LoggerComUsuario`, `VerificaRole`/`VerificaSuperAdmin`, `LimitarPorIP`) e a ordem em que entram na cadeia de uma requisição |
| [`ENDPOINTS.md`](./ENDPOINTS.md) | Todas as rotas da API por domínio: método, path, nível de autorização, request/response, erros notáveis |
| [`MODELO_DE_DADOS.md`](./MODELO_DE_DADOS.md) | Schema final consolidado do Postgres (tabelas, colunas, FKs, triggers), diagrama de relações |
| [`FLUXOS_DE_NEGOCIO.md`](./FLUXOS_DE_NEGOCIO.md) | Controle de estoque por lote (FEFO), entrega/devolução/troca, assinatura digital + token de auditoria, geração de PDF, importação via XLSX |
| [`BACKUP.md`](./BACKUP.md) | Subcomando `./main backup-banco`: dump do Postgres enviado pro bucket Cloudflare R2, variáveis `R2_*`, como testar via Docker e como restaurar |

## Documentação gerada (não editar à mão)

`docs.go`, `swagger.json` e `swagger.yaml` são gerados por `swag init` a partir das anotações
`// @Summary` / `// @Router` nos controllers — servidos em `/swagger/*any`. Regenere após alterar
rotas ou assinaturas de handler.
