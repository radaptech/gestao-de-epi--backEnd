# Fluxos de Negócio — SGEPI Backend

O core do domínio: como o estoque de EPIs é controlado por lote, como entrega/devolução consomem e
repõem esse estoque, e a mecânica de assinatura digital + token de auditoria comum aos dois fluxos.
Schema envolvido: [`MODELO_DE_DADOS.md`](./MODELO_DE_DADOS.md). Endpoints: [`ENDPOINTS.md`](./ENDPOINTS.md).

---

## Estoque: controle por lote

**O saldo não é um contador único por EPI.** Cada lote (uma linha de `entrada_epi_item`, vinculada
a uma NF em `entrada_nf`) tem seu próprio `quantidade_atual`, que só é abatido/reposto naquele lote
específico — nunca existe um "somatório global" persistido, ele é calculado em tempo real
(`SUM(quantidade_atual)`) quando necessário.

### Entrada (`EntradaService.Adicionar`)

Cria `entrada_nf` (cabeçalho: fornecedor, número/série da NF) + N `entrada_epi_item` (um por
lote/EPI/tamanho) **na mesma transação**. `quantidade_atual` nasce igual a `quantidade`. Validações
no service antes de tocar o banco:
- `data_emissao` não pode ser futura.
- `data_validade >= data_fabricacao` por item.

### Entrega (`EntregaService.RegistrarEntrega`) — FEFO com lock de linha

Para cada item pedido (`model.ItemParaInserir`: EPI + tamanho + quantidade):

1. `ListarEntregasDisponiveis` (query `ListarLotesParaConsumo`) busca os lotes candidatos:
   ```sql
   WHERE tenant_id = $1 AND id_epi = $2 AND id_tamanho = $3
     AND quantidade_atual > 0 AND data_validade >= CURRENT_DATE AND ativo = TRUE
   ORDER BY data_validade ASC
   FOR UPDATE
   ```
   `ORDER BY data_validade ASC` é o **FEFO** (First Expired, First Out) — sempre consome primeiro o
   lote que vence mais cedo. `FOR UPDATE` trava essas linhas dentro da transação, para que duas
   entregas concorrentes do mesmo EPI não leiam o mesmo saldo e o abatam em duplicidade (race
   condition clássica de estoque).
2. Se `len(lotes) == 0` → erro "estoque zerado ou validade vencida" imediatamente, sem tentar os
   outros itens do pedido.
3. Percorre os lotes em ordem; para cada um, `qtdAbater = min(lote.QuantidadeAtual,
   quantidadeNecessaria)`, insere uma linha em `epis_entregues` (`AdicionarEntregaItem`) apontando
   para aquele lote exato (`IDEntradaItem`) e chama `AbaterEstoqueLote`
   (`quantidade_atual -= qtdAbater`). Segue para o próximo lote se ainda sobrar quantidade.
4. Se depois de percorrer todos os lotes disponíveis `quantidadeNecessaria > 0`, a transação inteira
   sofre `rollback` e o erro retornado contém a string `"estoque insuficiente"` — é assim que o
   controller identifica esse caso específico para responder `422`.

**Um item pedido pode virar várias linhas em `epis_entregues`** se precisar consumir mais de um
lote para fechar a quantidade.

### Cancelamento de entrega (`EntregaService.RegistrarCancelamento`)

Marca a `entrega_epi` e os itens (`epis_entregues`) como cancelados/inativos e chama
`ReporEstoqueEntrada` (query `ReporEstoqueLote`) para devolver **exatamente** a quantidade
consumida ao **lote exato de origem** (`IDEntradaItem` gravado no momento da entrega) — não é uma
reposição genérica no "lote mais recente", é a reversão precisa da operação original.

### Devolução (`DevolucaoService.SalvarDevolucao`)

1. **Valida saldo em posse**: `ConsultaSaldo` (query `ConsultarSaldoEpiFuncionario`) calcula
   `entregue − devolvido` para aquele funcionário/EPI/tamanho. Se
   `quantidade_a_devolver > saldo_atual`, bloqueia com 422 antes de tocar em qualquer estoque.
2. **Verifica o motivo** (`MotivoDevolucao.Descarte`, coluna real `gera_descarte`):
   - Se **é descarte** (ex.: EPI danificado, vencido): o item **não volta** ao estoque. Só a linha
     de `devolucao` é registrada.
   - Se **não é descarte**: repõe a quantidade nos lotes existentes daquele EPI/tamanho, respeitando
     o espaço livre de cada um.
3. **Reposição por lote, do mais recente para o mais antigo** — diferente do FEFO da entrega. Query
   `ListarLotesParaRepor`:
   ```sql
   WHERE tenant_id = $1 AND id_epi = $2 AND id_tamanho = $3
     AND ativo = TRUE AND quantidade_atual < quantidade
   ORDER BY id DESC  -- do lote mais recente para o mais antigo
   ```
   Para cada lote candidato, `espaçoNoLote = lote.Quantidade − lote.QuantidadeAtual`; deixa
   `min(qtdRestante, espaçoNoLote)` naquele lote e segue para o próximo se sobrar quantidade. Se no
   fim ainda sobrar quantidade a repor (mais do que o histórico de compra do EPI comporta), é
   tratado como **anomalia de estoque** — erro explícito, não silenciado.
4. **Troca** (`houve_troca = true`): dentro da **mesma transação** (`qtx`), chama diretamente
   `EntregaService.RegistrarEntrega` para o EPI/tamanho/quantidade novos, vinculando a nova entrega
   à devolução via `entrega_epi.idtroca`. Se a nova entrega falhar (ex.: estoque insuficiente do EPI
   novo), a transação inteira — devolução incluída — sofre rollback.

### Cancelamento de devolução (`DevolucaoService.CancelarDevolucao`)

Reverte a devolução: chama `ReporEstoqueLote` para tirar de volta o que havia sido reposto, e marca
a devolução como cancelada. Se a devolução tinha gerado uma troca, a entrega vinculada também
precisa ser desfeita.

---

## Assinatura digital e token de auditoria

Entregas e devoluções exigem assinatura do funcionário. O fluxo é idêntico nos dois controllers,
antes de chamar o service:

1. **Gerar token de auditoria** — SHA-256 de `NOME|FUNÇÃO|DEPARTAMENTO|YYYY-MM-DD` (tudo em
   maiúsculo), truncado em 16 caracteres:
   - `helper.GerarTokenAuditoria(...)` → prefixo `ENT-`, para entregas.
   - `helper.GerarTokenDevolucao(...)` → prefixo `DEVO-`, para devoluções.
   Guardado em `entrega_epi.token_validacao` / `devolucao.token_validacao` — serve como prova de
   integridade do registro (mesmos dados de entrada sempre geram o mesmo token).
2. **Upload da assinatura** — `helper.UploadAssinaturaSupabase(base64, token, pasta)`: separa o
   prefixo `data:image/png;base64,` se vier junto, decodifica o base64, sobe como PNG no bucket do
   Supabase (nome do arquivo: `<pasta>/<token>_<unix timestamp>.png`) e devolve a **URL pública**.
3. O controller substitui `input.Assinatura_Digital` (que chegou em base64 do front) pela URL
   pública antes de chamar o service. **O banco só guarda a URL — nunca o base64 da imagem.**

## Geração de PDF (ficha de entrega / devolução)

`internal/helper/pdfHelperEntrega.go` e `pdfHelperDevolucao.go`, com **maroto/v2**: montam a ficha
com os dados da operação, código de barras e a imagem da assinatura (baixada da URL pública do
Supabase e rotacionada quando necessário). Endpoints:
- `GET /api/gerencial/{matricula}/ficha-pdf/{id}` — entrega.
- `GET /api/gerencial/devolucoes/{id}/pdf` — devolução.

## Matrícula do funcionário

Gerada por **trigger no Postgres** (`trigger_gerar_matricula`), não pela aplicação — deixe o campo
nulo no INSERT. Ver [`MODELO_DE_DADOS.md`](./MODELO_DE_DADOS.md#funcionario) para a lógica exata do
trigger.

## Limites de plano

`UsuarioService.Registrar` e `FuncionarioService` (no cadastro) comparam a contagem atual
(`TotalDeUsuario` / total de funcionários do tenant) contra `planos.limite_usuarios` /
`planos.limite_funcionarios`. Se excedido, retornam `helper.ErrLimiteExcedido` → HTTP `403`. Um
limite `NULL` no plano significa ilimitado.

## Importação via XLSX

Departamentos, funções e fornecedores aceitam upload de planilha (`multipart/form-data`, campo
`file`), lidos com **excelize** diretamente no controller (não no service):

| Recurso | Handler | Cabeçalho esperado (coluna A) | Comportamento em erro/duplicidade |
|---|---|---|---|
| Departamento | `ImportDepartamentoXLSX` | "Nome do Departamento" | Sem transação única — aborta no primeiro erro (`409`/`500`), linhas já inseridas permanecem |
| Função | `ImportarFuncaoXLSX` | "Nome da Função" (+ nome do departamento na coluna B) | Duplicadas/conflitos são **ignoradas e contadas** (`ignorados`), não abortam a importação |
| Fornecedor | `ImportFornecedor` | "Razão Social" (+ Nome Fantasia, CNPJ, Inscrição Estadual opcional) | Duplicados ignorados silenciosamente; ⚠️ o campo `total` da resposta sempre reporta `0` (bug conhecido: conta um slice local que nunca é populado) |

Em todos os três, a leitura abre a primeira aba e pula linhas até encontrar a linha de cabeçalho
esperada — planilhas com cabeçalho em posição diferente da esperada silenciosamente não importam
nada (0 linhas processadas), sem erro explícito de "cabeçalho não encontrado".

## Soft delete

Quase nada é apagado de verdade — ver a seção de nomenclatura em
[`MODELO_DE_DADOS.md`](./MODELO_DE_DADOS.md#observações-gerais-sobre-o-schema) para os diferentes
pares de coluna usados (`ativo`/`deletado_em` vs. `ativo`/`cancelada_em` vs. `ativo`/`cancelado_em`).
Listagens filtram pelo par correspondente da tabela; vários filtros de listagem expõem um booleano
`cancelados`/`canceladas` para consultar o oposto (ex.: `FiltroEntregas.Canceladas`).
