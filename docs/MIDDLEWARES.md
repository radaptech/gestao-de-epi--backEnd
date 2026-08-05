# Middlewares — SGEPI Backend

Todos vivem em `middleware/`. Este doc explica o que cada um faz, em que ordem entram na cadeia de
uma requisição e por quê.

---

## Ordem de execução

```
router.Use(...)                              → global, TODAS as rotas (inclusive /swagger)
  1. CorsConfig()
  2. SecurityHeaders()

/api/painel, /api/master (grupos separados)
  1. AutenticacaoJWT()
  2. VerificaSuperAdmin()

/api (grupo principal)
  ├─ sem middleware extra: /login, /logout, /esqueci-minha-senha, /redefinir-senha
  │    mas o GRUPO "api" já tem:
  1. TenantMiddleware(queries)              → aplicado ANTES do bloco de login/logout
  ── (rotas de login/logout ficam expostas só com Tenant, sem exigir JWT) ──
  2. AutenticacaoJWT()                      → aplicado depois, cascateando para tudo daqui pra baixo
  3. LoggerComUsuario()

  /api/gerencial (subgrupo de /api)
  4. VerificaRole("admin")
```

**Importante sobre `gin.RouterGroup`**: `api.Use(...)` é cascata — só afeta as rotas registradas
**depois** da chamada. É por isso que `POST /api/login` só passa pelo `TenantMiddleware` (registrado
antes do `Use(AutenticacaoJWT(), LoggerComUsuario())`), enquanto `GET /api/me` em diante passa pelos
dois. Ver `internal/routers/rotasHttp.go` para a ordem exata de registro.

---

## `CorsConfig()` — `middleware/cors.go`

Libera origem por `AllowOriginFunc`:
- `http://localhost` e qualquer porta (`http://localhost:*`)
- qualquer subdomínio `*.localhost` (dev, via Traefik)
- `https://radaptech.com.br` e subdomínios `*.radaptech.com.br` (produção/homologação)

`AllowCredentials: true` — obrigatório para o cookie HttpOnly de sessão funcionar cross-origin.
`AllowHeaders` inclui várias variações de capitalização de `X-Tenant-ID` (o header HTTP é
case-insensitive no protocolo, mas alguns proxies/browsers são estritos na whitelist do preflight).
`OptionsResponseStatusCode: 204` para o preflight.

**Um domínio novo (ex.: um novo cliente com domínio próprio) precisa ser adicionado aqui.**

## `SecurityHeaders()` — `middleware/SecurityHeaders.go`

Aplica em toda resposta:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Content-Security-Policy: default-src 'self'`

Roda até na UI do Swagger (`/swagger/*any`), já que é `router.Use` global — o CSP restritivo pode
quebrar recursos externos que a UI do Swagger tentasse carregar (hoje não tenta).

HSTS está comentado no código — só ative se o ambiente for HTTPS-only, senão bloqueia `localhost`.

## `TenantMiddleware(queries)` — `middleware/TenantId.go`

Resolve **qual empresa** está fazendo a requisição, a partir do header `X-Tenant-ID` (subdomínio,
não ID numérico):

1. Lê e valida que o header não está vazio → `400` se ausente.
2. Rejeita `www` e `api` como subdomínio (reservados) → `403`.
3. `queries.GetTenantBySubdomain(ctx, subdominio)` → `404` se a empresa não existir, `500` em outro
   erro de banco.
4. `c.Set("tenantId", empresa.ID)`.

Helper de leitura para os controllers: `middleware.GetTenantID(c) (int32, bool)`.

Recebe `*repository.Queries` (não o service) porque é o único middleware que consulta o banco
diretamente — é instanciado uma vez em `ConfigurarRotas` a partir do pool.

## `AutenticacaoJWT()` — `middleware/MiddJtw.go`

Extrai e valida o JWT, nessa ordem de preferência:

1. **Cookie `token`** (HttpOnly, setado no login) — checado primeiro.
2. Se não houver cookie, cai para o header `Authorization: Bearer <token>`.
3. Sem nenhum dos dois → `401` ("Token ausente ou formato inválido").

Valida a assinatura HMAC com `JWT_SECRET` (a secret é lida uma única vez, fora do handler, na
construção do middleware — não a cada requisição). Faz `panic` no boot se `JWT_SECRET` estiver
vazia — falha rápido em vez de rodar inseguro.

Em caso de token válido, extrai das claims e popula o contexto do Gin:
- `sub` → `c.Set("userId", int32(...))` — **obrigatório**; sem ele, `401`.
- `role` → `c.Set("user_role", ...)`.
- `tenantId` → `c.Set("user_TenantId", ...)`.
- `nome` → `c.Set("user_nome", ...)`.

Note que esse `tenantId` das claims do JWT (`user_TenantId`) é **distinto** do `tenantId` setado
pelo `TenantMiddleware` a partir do header `X-Tenant-ID` — o segundo é o que os controllers de
domínio usam (`middleware.GetTenantID`) para filtrar dados; o primeiro fica disponível caso algum
handler precise confirmar que o usuário logado pertence ao tenant do header.

## `LoggerComUsuario()` — `middleware/logs.go`

Roda depois do handler (`c.Next()` primeiro). Se `status >= 400`, loga
`⚠️  ERRO | status | método | path | User:<id ou Anonimo>`. Não loga requisições bem-sucedidas —
é log de erro, não access log completo.

## `VerificaRole(rolePermitida string)` e `VerificaSuperAdmin()` — `middleware/roles.go`

Comparação simples de string contra `c.GetString("user_role")` (setado pelo `AutenticacaoJWT`):

- `VerificaRole("admin")` — usado no grupo `/api/gerencial`. Compara **igualdade exata**; não há
  hierarquia de papéis aqui (um `super_admin` autenticado via `/api` normal, se existisse essa
  combinação, **não** passaria por `VerificaRole("admin")`, pois a comparação não é `>=`).
- `VerificaSuperAdmin()` — usado nos grupos `/api/painel` e `/api/master`. Só aceita `role ==
  "super_admin"`.

Resposta de erro em ambos: `403`, `{"erro": "..."}"` — note a chave `"erro"` (não `"error"` como no
resto da API), inconsistência a se ter em mente ao tratar erro no front.

## `LimitarPorIP(taxa rate.Limit, burst int)` — `middleware/RateLimit.go`

Token bucket **por IP**, implementado com `golang.org/x/time/rate`, um `*rate.Limiter` por IP num
mapa protegido por mutex. Uma goroutine de fundo (`limparInativos`) roda a cada 5 minutos e remove
do mapa qualquer IP inativo há mais de 10 minutos, para não vazar memória.

Usado pontualmente nas rotas mais sensíveis a força bruta/abuso, **não** globalmente:

| Rota | Taxa | Burst |
|---|---|---|
| `POST /api/login` | 1 tentativa a cada 12s | 5 de imediato |
| `POST /api/esqueci-minha-senha` | 1 a cada minuto | 3 de imediato |

Depende de `router.SetTrustedProxies(...)` estar configurado corretamente em `main.go` — sem isso,
`c.ClientIP()` pode confiar em um `X-Forwarded-For` forjado pelo próprio cliente, e o rate limit por
IP vira decorativo.

## `GetUserID(c)` — `middleware/userId.go`

Não é um middleware, é só o helper de leitura tipada do `userId` setado pelo `AutenticacaoJWT`
(`c.Get("userId")` → `int32`). Usado nos controllers para auditoria (`id_usuario_criacao`,
`id_usuario_cancelamento`, etc.).
