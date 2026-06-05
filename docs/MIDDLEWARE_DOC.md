# Documentação dos Middlewares

Este documento descreve detalhadamente cada middleware presente no diretório `middleware/` do projeto SGE (Sistema de Gestão de EPI).

---

## 1. MiddJtw.go (Autenticação JWT)

**Propósito:** Validar tokens JWT nas requisições HTTP, extrair claims e armazenar informações do usuário no contexto do Gin.

**Como funciona:**
- Lê a variável de ambiente `JWT_SECRET` para assinatura/verificação do token.
- Espera o token no header `Authorization` no formato `Bearer <token>`.
- Valida a assinatura, a expiração e a integridade do token.
- Extrai os claims: `sub` (ID do usuário), `role`, `tenantId`, `nome`.
- Armazena esses valores no contexto do Gin usando `ctx.Set()` para que handlers subsequentes possam acessá-los via `ctx.Get()`.

**Variáveis de ambiente:**
- `JWT_SECRET`: Chave secreta usada para assinar e verificar os tokens JWT. **Obrigatória.** Se vazia, o middleware causa um panic na inicialização.

**Fluxo de retorno:**
- Se o token estiver ausente ou mal formatado: retorna `401 Unauthorized` com JSON `{ "error": "Acesso negado: Token ausente ou formato inválido" }`.
- Se o token for inválido (assinatura incorreta, expirado, etc.): retorna `401 Unauthorized` com JSON `{ "error": "Token inválido ou expirado" }`.
- Se não conseguir converter os claims: retorna `401 Unauthorized` com JSON `{ "error": "Falha ao processar permissões" }`.
- Se o claim `sub` (ID do usuário) não estiver presente ou não for numérico: retorna `401 Unauthorized` com JSON `{ "error": "Token não contém identificação do usuário" }`.
- Em caso de sucesso: chama `ctx.Next()` para continuar a cadeia de handlers.

**Exemplo de uso no router:**
```go
r.Use(middleware.AutenticacaoJWT())
```

**Observações importantes:**
- O middleware deve ser colocado após middlewares que possam precisar do contexto (como CORS) mas antes dos handlers que exigem autenticação.
- As informações do usuário ficam disponíveis no contexto com as chaves: `userId` (int32), `user_role` (string), `user_TenantId` (int32), `user_nome` (string).
- O middleware não verifica permissões de role ou tenant; isso é feito por outros middlewares (como `VerificaRole`).

---

## 2. SecurityHeaders.go

**Propósito:** Adicionar headers de segurança HTTP em todas as respostas para proteger contra ataques comuns (XSS, clickjacking, MIME sniffing, etc.).

**Headers adicionados:**
- `X-Content-Type-Options: nosniff` – impede que o navegador tente adivinhar o tipo de conteúdo (MIME sniffing).
- `X-Frame-Options: DENY` – impede que a página seja exibida em um `<frame>`, `<iframe>`, `<embed>` ou `<object>`, protegendo contra clickjacking.
- `Content-Security-Policy: default-src 'self'` – restringe o carregamento de recursos (scripts, estilos, imagens, etc.) apenas ao mesmo origem, reduzindo risco de XSS e injeção de dados.
- `Strict-Transport-Security` (HSTS): **comentado por padrão**. Deve ser ativado apenas em produção com HTTPS. O valor sugerido é `max-age=31536000; includeSubDomains` (1 ano).

**Como usar:**
Basta registrar o middleware no router:
```go
r.Use(middleware.SecurityHeaders())
```

**Observações:**
- O header HSTS está comentado porque, se ativado em ambiente de desenvolvimento (HTTP) ou sem certificado válido, pode bloquear o acesso ao localhost. Ative-o apenas quando estiver usando HTTPS em produção.
- Esses headers são adicionados em **todas** as respostas, incluindo erros (4xx, 5xx).

---

## 3. TenantId.go

**Propósito:** Identificar o tenant (empresa) associado à requisição com base no subdomínio fornecido no header `X-Tenant-ID` (ou variantes de case) e buscar seu ID no banco de dados.

**Como funciona:**
1. Lê o header `X-Tenant-ID` (case-insensitive, aceita variações como `X-tenant-id`, `X-tenant-ID` devido à configuração do CORS).
2. Se o header estiver vazio, retorna `400 Bad Request` com `{ "error": "Tenant não informado (X-Tenant-ID ausente)" }`.
3. Ignora valores reservados `www` e `api` (retorna `403 Forbidden` com `{ "error": "Subdomínio reservado" }`).
4. Usa o repositório `repository.Queries` (gerado pelo sqlc) para chamar `GetTenantBySubdomain(ctx, subdominio)`.
5. Se não encontrar o tenant: retorna `404 Not Found` com `{ "error": "Empresa não encontrada" }`.
6. Se ocorrer outro erro no banco: retorna `500 Internal Server Error` com `{ "error": "Erro interno ao validar empresa" }`.
7. Em caso de sucesso: armazena o ID do tenant no contexto do Gin com a chave `"tenantId"` (constante `TenantId = "tenantId"`).
8. Chama `ctx.Next()` para continuar.

**Helper function:**
- `GetTenantID(c *gin.Context) (int32, bool)`: função auxiliar para obter o ID do tenant de forma tipada a partir do contexto. Retorna o valor e um booleano indicando se existe.

**Dependências:**
- Precisa de uma instância inicializada de `*repository.Queries` (do sqlc) para ser passado ao middleware na criação.

**Exemplo de uso no router:**
```go
tenantRepo := repository.NewQueries(db) // ou como seu código obtém o Queries
r.Use(middleware.TenantMiddleware(tenantRepo))
```

**Observações:**
- O middleware pressupõe que o banco de dados contém uma tabela de tenants com pelo menos os campos `id` e `subdomain`.
- O header `X-Tenant-ID` deve ser enviado pelo cliente (geralmente o frontend) em cada requisição.
- O middleware deve vir **depois** do middleware de CORS (para que o header seja permitido) e **depois** do middleware de autenticação JWT se precisar do `user_TenantId` do token para validação adicional (este middleware não faz aquela validação; ele apenas busca pelo subdomínio no header).

---

## 4. cors.go

**Propósito:** Configurar o middleware CORS (Cross-Origin Resource Sharing) do Gin para controlar quais origens podem acessar a API, quais métodos e headers são permitidos, e se credenciais (cookies, auth headers) podem ser incluídos.

**Configuração detalhada (via `cors.New(cors.Config{...})`):**

- `AllowOriginFunc`: função que determina se uma origem é permitida.
  - Origens explicitamente permitidas (exemplos; ajuste conforme seu ambiente):
    - `http://localhost:3000` (frontend local em desenvolvimento)
    - `https://sgepi-homologacao.radaptech.com.br` (homologação)
    - `http://teste.localhost`
    - `http://painel.localhost`
    - `https://radaptech.com.br`
    - `http://localhost`
  - Também permite qualquer subdomínio de `.radaptech.com.br` (ex: `app.radaptech.com.br`).
  - Todas as outras origens são bloqueadas.

- `AllowMethods`: `["POST", "PUT", "GET", "PATCH", "DELETE", "OPTIONS"]` – métodos HTTP permitidos.

- `AllowHeaders`: headers que o cliente pode enviar em requisições CORS.
  - Inclui: `Origin`, `Content-Type`, `Accept`, `Authorization`, `X-Requested-With`, e variantes de case do `X-Tenant-ID` (`X-tenant-id`, `X-tenant-ID`) para garantir que o header do tenant seja permitido independentemente da capitalização.

- `ExposeHeaders`: headers que o frontend pode acessar na resposta (além dos simples como `Cache-Control`, `Content-Language`, etc.).
  - `Content-Length` e `Content-Disposition` – útil para downloads de arquivos, onde o frontend precisa saber o nome do arquivo via `Content-Disposition`.

- `AllowCredentials: true` – permite que cookies e headers de autorização (como `Authorization`) sejam incluídos em requisições cross-origin.

- `MaxAge: 12 * time.Hour` – por quanto tempo o navegador pode cachear o resultado de uma preflight request (OPTIONS).

**Como usar:**
```go
r.Use(middleware.CorsConfig())
```

**Observações:**
- O middleware CORS deve ser registrado **antes** de qualquer outro middleware que precise ler headers como `Authorization` ou `X-Tenant-ID`, pois ele trata o preflight (OPTIONS) e permite que os headers sejam enviados na requisição real.
- A lista de origens permitidas deve ser mantida atualizada conforme os ambientes (desenvolvimento, homologação, produção) forem adicionados ou removidos.
- Em produção, evite usar curingas como `*` quando `AllowCredentials` estiver verdadeiro (não é permitido pela especificação CORS).

---

## 5. logs.go

**Propósito:** Fazer logging de requisições que resultaram em erros (status >= 400), incluindo o ID do usuário (se autenticado) para facilitar a auditoria e depuração.

**Como funciona:**
- Chama `ctx.Next()` para deixar o handler processar a requisição.
- Após o handler terminar, obtém o status da resposta via `c.Writer.Status()`.
- Se o status for maior ou igual a 400 (erro de cliente ou servidor):
  - Tenta obter o usuário do contexto pela chave `"userId"` (definido pelo middleware JWT).
  - Se existir, formata como `User:<id>`; caso contrário, usa `"Anonimo"`.
  - Imprime um log no formato:
    ```
    ⚠️  ERRO | <status> | <method> | <path> | <userLog>
    ```
    Exemplo: `⚠️  ERRO | 401 | GET | /api/epi | User:123`

**Observações:**
- Só logs erros; requisições bem-sucedidas (status < 400) são ignoradas para evitar poluição do log.
- O log usa o pacote padrão `log` do Go, que escreve em stderr (geralmente capturado pelo sistema de orquestração como Docker/Kubernetes).
- Não inclui informações sensíveis como headers ou corpo da requisição.
- Para logging mais avançado (níveis, formatos JSON, etc.), considere usar uma biblioteca especializada como `zap` ou `zerolog`.

**Como usar:**
```go
r.Use(middleware.LoggerComUsuario())
```

---

## 6. roles.go

**Propósito:** Verificar se o usuário autenticado possui uma determinada role ou se é super_admin, concedendo ou negando acesso conforme o caso.

### Funções disponíveis:

#### `VerificaRole(rolePermitida string) gin.HandlerFunc`
- **Propósito:** Verifica se a role do usuário (armazenada no contexto pela chave `"user_role"` pelo middleware JWT) corresponde exatamente à role permitida.
- **Como funciona:**
  - Obtém `roleAtual := ctx.GetString("user_role")`.
  - Se `roleAtual` for diferente de `rolePermitida`, retorna `403 Forbidden` com JSON `{ "erro": "Acesso negado: privilégios insuficientes" }` e chama `ctx.Abort()` para interromper a cadeia.
  - Caso contrário, chama `ctx.Next()`.
- **Uso típico:** proteger endpoints que só podem ser acessados por determinado tipo de usuário (ex: apenas `admin`, apenas `operador`).

#### `VerificaSuperAdmin() gin.HandlerFunc`
- **Propósito:** Verificar se o usuário tem a role `super_admin` (acesso total ao sistema).
- **Como funciona:**
  - Obtém `roleAtual := ctx.GetString("user_role")`.
  - Se `roleAtual` não for exatamente `"super_admin"`, retorna `403 Forbidden` com JSON `{ "erro": "Acesso negado: área restrita à administração global do SGE" }`.
  - Caso contrário, chama `ctx.Next()`.
- **Uso típico:** proteger endpoints de configuração global, gerenciamento de tenants, usuários master, etc.

**Observações:**
- Ambos os middleware pressupõem que o middleware de autenticação JWT foi executado anteriormente e definiu a chave `"user_role"` no contexto.
- A comparação é exata e sensível a case (as roles são esperadas em minúsculas, como `"admin"`, `"operador"`, `"super_admin"` conforme definido no token).
- Para múltiplas roles permitidas (ex: admin ou operador), seria necessário criar um middleware customizado ou usar esta função várias vezes com lógica OR (não fornecido aqui).
- Esses middleware devem ser colocados após o middleware JWT e após o middleware de tenant (se a verificação de role depender do tenant).

---

## Histórico e Manutenção

Este documento foi criado com base na análise dos arquivos fonte em `/home/davif/back/sge/gestao-de-epi--backEnd/middleware/` na data de 2026-06-04.

Qualquer alteração nos arquivos de middleware deve ser refletida neste documento para manter a consistência.

Recomenda-se revisar este documento sempre que:
- Novos middlewares forem adicionados.
- Middlewares existentes tiverem sua assinatura ou comportamento alterado.
- Variáveis de ambiente ou dependências forem modificadas.

--- 
*Fim da documentação.*