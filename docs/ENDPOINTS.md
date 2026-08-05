# Endpoints — SGEPI Backend

Referência prática das rotas HTTP registradas em `internal/routers/rotasHttp.go` (fonte de verdade
usada aqui — algumas anotações `@Router` do Swagger nos controllers estão desatualizadas em relação
à rota real registrada; onde isso acontece, uma nota é deixada explicitamente). Para o schema
completo de request/response gerado formalmente, use o Swagger em `/swagger/*any` — este documento
prioriza contexto de negócio e pegadinhas que o Swagger não mostra.

## Padrões comuns a quase todos os endpoints

- **Tenant**: resolvido pelo `TenantMiddleware` a partir do header `X-Tenant-ID` (subdomínio) e lido
  no controller via `middleware.GetTenantID(ctx)`. Rotas de `/api/painel` e `/api/master` não têm
  tenant (são globais, só `super_admin`).
- **Auth**: cookie HttpOnly `token` (preferencial) ou header `Authorization: Bearer <token>`.
- **Paginação**: query string `pagina` / `quantidade`, normalizada no service (limite máx. 100 por
  página na maioria; defaults variam por endpoint — ver tabelas abaixo).
- **Erro padrão**: `{"error": "<mensagem amigável>", "detalhes": "<err.Error()>"}`. Existem
  exceções pontuais que usam a chave `"erro"` em vez de `"error"` (`PlanosController.SalvarPlano`,
  `EmpresaController.EditarEmpresa`, e os dois middlewares de role) — trate ambas no front se for
  parsear a chave.
- **Grupos de auth**: Colaborador/Admin = qualquer usuário autenticado sob `/api`; Admin = também
  precisa passar por `VerificaRole("admin")` em `/api/gerencial`; Super admin = `/api/painel` ou
  `/api/master`.

---

## Autenticação / Usuários

| Método | Path | Auth | Request | Response (sucesso) | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| POST | `/api/login` | Tenant-only (rate limit: 5 burst, 1/12s por IP) | `model.LoginInput` (email, senha) | `{usuario: {id, nome, email, role}}` + cookie HttpOnly `token` | Autentica e seta cookie JWT | 400 dados inválidos; 401 email/senha incorretos; 500 |
| POST | `/api/logout` | Tenant-only | — | `{message}` | Limpa o cookie `token` | — |
| POST | `/api/esqueci-minha-senha` | Tenant-only (rate limit: 3 burst, 1/min por IP) | `model.RecuperaLogin` | 200 vazio | Dispara e-mail de recuperação de senha (Resend) | 400; 500 |
| POST | `/api/redefinir-senha` | Tenant-only | `model.RedefinirSenha` | `{mensagem}` | Redefine senha via token do link de e-mail | 400 link inválido/expirado/empresa errada; 500 |
| GET | `/api/me` | Colaborador/Admin | — | `{id, nome, email, role}` | Perfil do usuário logado | 401; 404 não encontrado |
| GET | `/api/gerencial/usuarios` | Admin | — | `[]model.UsuarioResponse` | Lista usuários do tenant | 500 |
| GET | `/api/master/dashboard/dados-usuarios` | Super admin | — | `[]model.UsuarioResponsePainel` | Lista usuários globais (painel master) | 500 |
| POST | `/api/master/dashboard/salvar-usuarios` | Super admin | `model.Usuario` (nome, email, senha, role, empresaId) | 201 `{mensagem}` | Cria usuário (qualquer tenant) | 409 duplicado; 403 `ErrLimiteExcedido` (limite de usuários do plano); 500 |
| PATCH | `/api/master/dashboard/editar/{id}` | Super admin | `model.EditarUsuarioRequest` | 204 | Edita dados de um usuário | 500 |
| PATCH | `/api/master/dashboard/usuario/{id}/status` | Super admin | `model.AlterarStatusRequest` | 204 | Ativa/inativa um usuário | 500 |

## Empresas / Planos (super_admin)

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| POST | `/api/painel/cadastrar-empresa` | Super admin | `model.EmpresaInserir` | 201 `{sucesso}` | Cadastra nova empresa (tenant) | 400; 500 |
| GET | `/api/master/dashboard/resumo` | Super admin | — | `model.ResumoDashboard` | Resumo global (dashboard master) | 500 |
| GET | `/api/master/dashboard/empresas-recentes` | Super admin | — | `[]model.EmpresaRecente` | Empresas cadastradas recentemente | 500 |
| GET | `/api/master/dashboard/dados-empresas` | Super admin | — | `[]model.Empresa` | Lista todas as empresas | 500 |
| PATCH | `/api/master/dashboard/{id}/empresa` | Super admin | `model.EditarEmpresaRequest` | 204 | Edita dados de uma empresa | 400 id inválido; 500 |
| GET | `/api/master/dashboard/planos` | Super admin | — | `[]model.Plano` | Lista planos disponíveis | 500 |
| POST | `/api/master/dashboard/cadastrar-planos` | Super admin | `model.Plano` | 201 `int32` (id do plano) | Cria um novo plano | 400; 500 |
| PATCH | `/api/master/dashboard/planos/{id}` | Super admin | `model.AtualizarPlanoParams` | 200 `{sucesso}` | Atualiza limites/dados de um plano | 400 id inválido; 500 |
| PATCH | `/api/master/dashboard/planos/{id}/status` | Super admin | `{status: string}` | 200 `{sucesso}` | Ativa/inativa um plano | 400 status ausente; 500 |

## Departamento

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/departamentos` | Colaborador/Admin | query `service.FiltroDepartamento` | `service.DepartamentoPaginado` | Lista paginada de departamentos | 400; 500 |
| POST | `/api/gerencial/cadastro-departamento` | Admin | `model.Departamento` | 201 `{mensagem, departamento}` | Cria departamento | 400; 409 já existe; 500 |
| POST | `/api/gerencial/importar-departamentos` | Admin | `multipart/form-data`, campo `file` (.xlsx) | 200 `{message, total}` | Importa departamentos em lote. Lê a 1ª aba, pula linhas até achar cabeçalho **"Nome do Departamento"** (coluna A) | 400 arquivo inválido/vazio; 401 sessão inválida; 409 duplicado (aborta no primeiro erro); 500 |
| PUT | `/api/gerencial/departamento/{id}` | Admin | `model.Departamento` | 200 `{sucesso}` | Atualiza nome do departamento | 400 nome curto (<2 letras) ou id inválido; 404 não encontrado; 500 |
| DELETE | `/api/gerencial/departamento/{id}` | Admin | — | 204 | Remove (soft-delete) departamento | 400 id inválido; 404 não encontrado; 500 |

## Função

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/funcoes` | Colaborador/Admin | query `service.FiltroFuncao` | `service.FuncaoPaginado` | Lista paginada de funções | 400; 500 |
| POST | `/api/gerencial/cadastro-funcao` | Admin | `model.Funcao` (nome, idDepartamento) | 200 `{mensagem}` | Cria função vinculada a um departamento | 422 duplicado; 409 conflito de integridade (departamento inexistente); 500 |
| POST | `/api/gerencial/importar-funcoes` | Admin | `multipart/form-data`, campo `file` (.xlsx) | 200 `{message, importados, ignorados}` | Importa funções em lote. Cabeçalho esperado **"Nome da Função"** (col. A) + nome do departamento (col. B); duplicadas/conflitos são ignoradas e contadas, não abortam a importação | 400 arquivo inválido/departamento não cadastrado; 401 sessão inválida; 500 |
| PUT | `/api/gerencial/funcao/{id}` | Admin | `model.Funcao` | 200 `{sucesso}` | Atualiza nome da função | 422 nome curto; 404 não encontrado; 500 |
| DELETE | `/api/gerencial/funcao/{id}` | Admin | — | 204 | Remove (soft-delete) função | 400 id inválido; 422 não encontrada; 500 |

## Funcionário

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/funcionarios` | Colaborador/Admin | query `service.FiltroFuncionario` | `service.FuncionarioPaginado` | Lista paginada de funcionários | 400; 500 |
| GET | `/api/funcionario/{matricula}` | Colaborador/Admin | path `matricula` | `model.Funcionario_Dto` | Busca funcionário por matrícula | 404 não encontrado; 500 |
| GET | `/api/funcionarios-dashbord` | Colaborador/Admin | — | `[]model.FuncionarioDashbord` | Resumo de funcionários para dashboard | 500 |
| GET | `/api/funcionario-estoque` | Colaborador/Admin | — | `[]model.FuncionarioCompletoDto` | Lista detalhada de funcionários (handler `FuncionarioCompleto`) | 500 |
| POST | `/api/gerencial/cadastro-funcionario` | Admin | `model.FuncionarioInserir` (nome, idDepartamento, idFuncao, cpf) | 201 `{mensagem}` | Cadastra funcionário. Matrícula é gerada por **trigger no Postgres**, não enviada no body | 409 duplicado; 403 `ErrLimiteExcedido` (limite do plano); 422 departamento/função inexistente; 500 |
| PATCH | `/api/gerencial/funcionario/{id}` | Admin | `model.UpdateFuncionarioRequest` | 200 `{sucesso}` | Atualiza dados completos do funcionário | 400; 404 não encontrado; 409 matrícula duplicada; 422 departamento/função inexistente; 500 |
| DELETE | `/api/gerencial/funcionario/{id}` | Admin | — | 204 | Remove (soft-delete) funcionário | 400 id inválido; 404 não encontrado; 500 |

## Tipo de Proteção

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/protecoes` | Colaborador/Admin | — | `[]model.TipoProtecaoDto` | Lista tipos de proteção | 500 |
| GET | `/api/protecao/{id}` | Colaborador/Admin | path `id` | `model.TipoProtecaoDto` | Busca tipo de proteção por ID | 400 id inválido; 404 não encontrado; 500 |
| POST | `/api/gerencial/cadastro-protecao` | Admin | `model.TipoProtecao` (nome) | 200 `{mensagem, protecao}` | Cria tipo de proteção | 409 duplicado; 500 |
| DELETE | `/api/gerencial/protecao/{id}` | Admin | — | 204 | Remove tipo de proteção | 400 id inválido; 404 não encontrado; 500 |

## Tamanho

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/tamanhos` | Colaborador/Admin | — | `[]model.TamanhoDto` | Lista todos os tamanhos | 500 |
| GET | `/api/tamanhos-id-epi/{id}` | Colaborador/Admin | path `id` do EPI | `[]model.TamanhoEntregaDto` | Tamanhos disponíveis para um EPI (tela de entrega) | 400 id inválido; 404 não encontrado; 500 |
| GET | `/api/tamanho/{id}` | Colaborador/Admin | path `id` | `model.TamanhoDto` | Busca um tamanho por ID | 400; 404 não encontrado; 500 |
| POST | `/api/gerencial/cadastro-tamanho` | Admin | `model.Tamanhos` (tamanho) | 201 `{mensagem}` | Cria tamanho | 409 duplicado; 500 |
| DELETE | `/api/gerencial/tamanho/{id}` | Admin | — | 204 | Remove (cancela) tamanho | 400; 404 não encontrado; 500 |

## EPI

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/epis` | Colaborador/Admin | query `service.EpiFiltro` | `service.EpiPaginado` | Lista paginada de EPIs | 400; 500 |
| GET | `/api/epi/{id}` | Colaborador/Admin | path `id` | `model.EpiDto` | Busca EPI por ID | 400; 422 não encontrado; 500 |
| GET | `/api/epis-dashbord` | Colaborador/Admin | — | `[]model.EpiDashBord` | Resumo de EPIs para dashboard | 500 |
| GET | `/api/funcionarios/{id}/epis` | Colaborador/Admin | path `id` do funcionário | `[]model.EpiDtoDevolucao` | EPIs em posse de um funcionário (tela de devolução) | 400; 500 |
| POST | `/api/gerencial/cadastro-epi` | Admin | `model.EpiInserir` | 200 `{mensagem}` | Cadastra EPI | 409 CA já registrado; 422 data inválida ou tamanho/proteção inexistente; 500 |
| PATCH | `/api/gerencial/epi/{id}` | Admin | `model.UpdateEpiInput` | 200 `{sucesso}` | Atualiza EPI existente | 409 CA duplicado; 422 não encontrado/data inválida/tamanho ou proteção inexistente; 500 |
| DELETE | `/api/gerencial/epi/{id}` | Admin | — | 204 | Remove (soft-delete) EPI | 400; 404 não encontrado; 500 |

## Fornecedor

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/fornecedores` | Colaborador/Admin | query `service.FiltroFornecedores` | `service.FornecedoresPaginados` | Lista paginada de fornecedores | 400; 500 |
| POST | `/api/gerencial/cadastro-fornecedores` | Admin | `model.FornecedorInserir` | 200 `{mensagem}` | Cadastra fornecedor | 409 CNPJ duplicado; 500 |
| POST | `/api/gerencial/importar-fornecedores` | Admin | `multipart/form-data`, campo `file` (.xlsx) | 200 `{message, total}` | Importa fornecedores. Cabeçalho esperado **"Razão Social"** (col. A) + Nome Fantasia (B) + CNPJ (C) + Inscrição Estadual opcional (D); duplicados são ignorados silenciosamente (não contam em `total`) | 400 arquivo inválido/vazio; 401 sessão inválida; 500 |
| PATCH | `/api/gerencial/fornecedor/{id}` | Admin | `model.FornecedorUpdate` | 200 `{sucesso}` | Atualiza fornecedor | 400; 409 CNPJ duplicado; 500 |
| DELETE | `/api/gerencial/fornecedor/{id}` | Admin | — | 204 | Cancela fornecedor | 400; 404 não encontrado; 500 |

## Entrada / Estoque

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/entradas` | Colaborador/Admin | query `service.FiltroEntradas` | `service.EntradaPaginada` | Lista paginada de entradas de NF | 400; 500 |
| GET | `/api/entradas-dashbord` | Colaborador/Admin | — | `[]model.EntradaDashbord` | Resumo de entradas para dashboard | 500 |
| GET | `/api/entradas-estoque` | Colaborador/Admin | — | `[]model.EntradaEstoqueDto` | Entradas relevantes para consulta de estoque | 500 |
| GET | `/api/quantidade-epi` | Colaborador/Admin | query `service.FiltroEstoqueAtual` | `service.EstoqueAtualPaginado` | Quantidade total em estoque por EPI/lote | 400; 500 |
| GET | `/api/saldo-epi` | Colaborador/Admin | query `service.FiltroEstoqueSaldo` | `service.EstoqueSaldoPaginado` | Saldo valorizado do estoque | 400; 500 |
| POST | `/api/gerencial/cadastrar-entrada` | Admin | `model.EntradaEpiInserir` (NF + itens) | 200 `{mensagem}` | Registra entrada de NF com N itens/lotes numa transação; grava `id_usuario_criacao` | 422 data de entrada/fabricação/validade inconsistente, NF duplicada, ou EPI/tamanho/fornecedor inexistente; 500 (⚠️ bug conhecido: em alguns caminhos de erro o handler não faz `return` após responder o JSON de erro, então o fluxo pode continuar) |
| DELETE | `/api/gerencial/entrada/{id}` | Admin | — | 204 | Cancela entrada (grava `id_usuario_cancelamento`) | 400; 404 não encontrado; 500 |

## Entrega

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/entregas` | Colaborador/Admin | query `service.FiltroEntregas` | `service.EntregaPaginada` | Lista paginada de entregas | 400; 500 |
| POST | `/api/cadastro-entregas` | Colaborador/Admin | `model.EntregaParaInserir` (idFuncionário, itens, assinatura em base64) | 200 `{mensagem}` | Gera token `ENT-...` → sobe assinatura pro Supabase → consome estoque FEFO por lote (`epis_entregues`) | 400; 422 funcionário/registro não encontrado, ou "estoque insuficiente"; 500 |
| GET | `/api/entregas-dashbord` | Colaborador/Admin | — | `[]model.EntregaDashbord` | Estatísticas de entregas p/ dashboard | 500 |
| GET | `/api/entrega-itens-dashbord` | Colaborador/Admin | — | `[]model.EntregaItensDashBord` | Estatísticas por item entregue p/ dashboard | 500 |
| DELETE | `/api/gerencial/entrega/{id}` | Admin | — | 204 | Cancela entrega e repõe estoque no lote exato de origem | 400; 404 não encontrado; 500 |
| GET | `/api/gerencial/{matricula}/ficha-pdf/{id}` | Admin | path `matricula`, `id` da entrega | PDF binário | Gera ficha de entrega em PDF (assinatura, código de barras) | 400 id inválido; 422 dados não encontrados; 500 |

## Devolução

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| POST | `/api/cadastro-devolucao` | Colaborador/Admin | `model.DevolucaoInserir` (idFuncionario, idEpi, idMotivo, quantidade, assinatura base64, opcional troca) | 200 `{mensagem}` | Gera token `DEVO-...` → sobe assinatura → valida saldo em posse → repõe estoque se motivo não for descarte; pode disparar nova entrega se `houveTroca` | 400; 422 funcionário/registro não encontrado; 500 |
| GET | `/api/devolucao` | Colaborador/Admin | — | `[]model.DevolucaoResponse` | Lista todas as devoluções do tenant (sem paginação) | 500 |
| GET | `/api/gerencial/devolucoes/{id}/pdf` | Admin | path `id` | PDF binário | Gera ficha de devolução em PDF | 400 id inválido; 422 não encontrado; 500 |

## Motivo de Devolução

| Método | Path | Auth | Request | Response | Descrição | Erros notáveis |
|---|---|---|---|---|---|---|
| GET | `/api/motivos` | Colaborador/Admin | — | `[]model.MotivoDevolucaoEpiDto` | Lista motivos de devolução (com flag de descarte) | 500 |
| POST | `/api/cadastrar-motivo-devolucao` | Colaborador/Admin (⚠️ não está sob `/gerencial` — qualquer colaborador autenticado pode cadastrar, não só admin) | `model.MotivoDevolucao` (motivo, ehDescarte) | 201 `model.MotivoDevolucaoEpiDto` | Cria motivo de devolução | 409 duplicado; 500 |

---

Para os fluxos de negócio por trás de Entrada/Entrega/Devolução (FEFO, controle por lote,
assinatura digital), ver [`FLUXOS_DE_NEGOCIO.md`](./FLUXOS_DE_NEGOCIO.md). Para os papéis/roles e a
autenticação em si, ver [`MIDDLEWARES.md`](./MIDDLEWARES.md).
