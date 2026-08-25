# Modelo de Dados — SGEPI Backend

Estado consolidado do schema **depois de aplicar todas as 65 migrações** em `database/migrate/`
(`000001` a `000033`, pares `.up.sql`/`.down.sql`). Este documento descreve a forma final das
tabelas, não o histórico migração a migração — para isso, leia os arquivos em `database/migrate/`
diretamente.

> `sqlc` gera o código de acesso em `database/repository/*.sql.go` a partir de
> `database/queries/*.sql`, usando este schema como referência. Nunca edite o `.sql.go` à mão.

---

## Diagrama de relações (FKs)

```
empresas (tenant) ─┬─ usuarios (tenant_id NULLABLE — super_admin é global, sem tenant)
                    ├─ planos ←── empresas.plano_id (ON DELETE SET NULL)
                    ├─ departamento ─┬─ funcao ─── funcionario ─┬─ entrega_epi
                    │                └─ funcionario              └─ devolucao
                    ├─ tipo_protecao ─── epi ─┬─ tamanhos_epis ─── tamanho
                    │                          └─ entrada_epi_item
                    ├─ fornecedores ─── entrada_nf ─── entrada_epi_item
                    ├─ entrega_epi ─── epis_entregues ─┬─ entrada_epi_item (lote de origem)
                    │                                   └─ epi
                    ├─ motivo_devolucao ─── devolucao
                    └─ devolucao ─(troca opcional)─ epi/tamanho novos
```

---

## Tabelas

### `empresas` (tenant)

| Coluna | Tipo | Notas |
|---|---|---|
| `id` | SERIAL PK | |
| `nome_fantasia`, `razao_social` | VARCHAR(100) | |
| `cnpj` | VARCHAR(20) UNIQUE | |
| `subdominio` | VARCHAR(50) UNIQUE | resolvido pelo `TenantMiddleware` a partir do header `X-Tenant-ID` |
| `criado_em` | TIMESTAMP | default now |
| `plano_id` | INT NULL → `planos(id)` | `ON DELETE SET NULL` |
| `status` | VARCHAR(20) NOT NULL DEFAULT `'Em teste'` | substituiu a antiga coluna `ativo` (dropada) |
| `vencimento` | DATE | |
| `observacoes` | TEXT | |
| `responsavel`, `email`, `telefone` | VARCHAR | dados de contato comercial |

Não tem coluna `mensalidade` — foi adicionada e depois removida; o valor cobrado vive em
`planos.mensalidade`.

### `usuarios`

| Coluna | Tipo | Notas |
|---|---|---|
| `id` | SERIAL PK | |
| `tenant_id` | INT **NULLABLE** → `empresas(id)` | nulo apenas para `super_admin` |
| `nome`, `email`, `senha_hash` | | senha via argon2id (`auth.HashPassword`) |
| `ativo` | BOOL DEFAULT TRUE | |
| `role` | VARCHAR(20) **sem default** | precisa ser enviado explicitamente ao criar — o `DEFAULT 'colaborador'` foi removido numa migração posterior |
| `ultimo_acesso` | TIMESTAMP NULL | |
| `token_recuperacao_senha` | TEXT NULL | |
| `token_expiracao` | TIMESTAMP NULL | |

`UNIQUE(tenant_id, email)`.

### `departamento`, `tipo_protecao`, `tamanho`

Mesmo formato para as três: `id` PK, `tenant_id`, um campo de nome (`nome` ou `tamanho`), `ativo`,
`deletado_em`. Índice único **parcial**: `UNIQUE(tenant_id, nome) WHERE deletado_em IS NULL` — o
soft delete libera o nome para reuso.

### `funcao`

`id`, `tenant_id`, `nome`, `iddepartamento → departamento(id)`, `ativo`, `deletado_em`.

> ⚠️ **Duas constraints de unicidade sobrepostas**: um índice parcial `(tenant_id, nome) WHERE
> deletado_em IS NULL` (migração 007) **e** uma `UNIQUE(tenant_id, nome, iddepartamento)` cheia,
> sem filtro de `deletado_em` (migração 033). A segunda é mais restritiva e nunca libera o par
> nome+departamento para reuso, mesmo com o registro antigo soft-deletado — provável inconsistência
> não intencional entre as duas migrações. Ao investigar erros `409`/`422` de "função duplicada"
> num nome que parece livre, checar as duas constraints.

### `epi`

`id`, `tenant_id`, `nome`, `fabricante`, `ca VARCHAR(20)`, `descricao`, `validade_ca DATE`,
`idtipoprotecao → tipo_protecao(id)`, `alerta_minimo INT`, `ativo`, `deletado_em`.
`UNIQUE(tenant_id, ca)` — o CA (Certificado de Aprovação) só precisa ser único dentro da empresa.

### `tamanhos_epis`

Tabela associativa `epi` ↔ `tamanho`: `id`, `tenant_id`, `idepi`, `idtamanho`, `ativo`,
`deletado_em`. `UNIQUE(idepi, idtamanho, tenant_id)`.

### `funcionario`

| Coluna | Tipo | Notas |
|---|---|---|
| `id` | SERIAL PK | |
| `tenant_id` | INT | |
| `nome` | | |
| `matricula` | **INT** | gerada por trigger (ver abaixo), não pela aplicação |
| `idfuncao → funcao(id)`, `iddepartamento → departamento(id)` | | |
| `ativo`, `deletado_em` | | |
| `cpf` | VARCHAR(11) NULLABLE | adicionado numa migração posterior |

`UNIQUE(tenant_id, matricula)`, `UNIQUE(tenant_id, cpf)`.

**Trigger `trigger_gerar_matricula`** (`BEFORE INSERT ... WHEN NEW.matricula IS NULL`), function
`gerar_matricula_por_tenant()`:
```sql
SELECT COALESCE(MAX(matricula), 0) + 1 FROM funcionario WHERE tenant_id = NEW.tenant_id
```
Sequencial **por tenant** — nunca gere matrícula na aplicação, deixe o campo nulo no INSERT.

### `fornecedores`

`id`, `tenant_id`, `razao_social`, `nome_fantasia`, `cnpj VARCHAR(14)`, `inscricao_estadual`,
`ativo BOOL DEFAULT TRUE`, `cancelado_em TIMESTAMP NULL`. `UNIQUE(tenant_id, cnpj)`.

> Nome de coluna fora do padrão: usa `cancelado_em`, não `deletado_em` como as tabelas de cadastro
> acima.

### `entrada_nf` (cabeçalho da nota fiscal)

| Coluna | Tipo | Notas |
|---|---|---|
| `id` | SERIAL PK | |
| `tenant_id` | INT | |
| `nota_fiscal_numero`, `nota_fiscal_serie DEFAULT '1'` | | |
| `data_emissao` | DATE | validada como não-futura no service |
| `data_registro` | TIMESTAMP DEFAULT now | |
| `idfornecedor → fornecedores(id)` NOT NULL | | substituiu uma coluna texto livre `fornecedor` |
| `id_usuario_criacao → usuarios(id)` NOT NULL | | auditoria |
| `id_usuario_cancelamento → usuarios(id)` NULL | | auditoria |
| `ativo`, `cancelada_em` | | |

`UNIQUE(tenant_id, idfornecedor, nota_fiscal_numero, nota_fiscal_serie)`.

### `entrada_epi_item` (o lote — fonte de verdade do saldo de estoque)

| Coluna | Tipo | Notas |
|---|---|---|
| `id` | SERIAL PK | |
| `tenant_id` | INT | |
| `entrada_nf_id → entrada_nf(id)` | `ON DELETE CASCADE` | |
| `id_epi → epi(id)`, `id_tamanho → tamanho(id)` | | |
| `quantidade` | INT | quantidade original do lote |
| `quantidade_atual` | INT | **saldo atual do lote** — decrementado a cada entrega, incrementado em cancelamento/devolução |
| `data_fabricacao`, `data_validade` | DATE | validado: `data_validade >= data_fabricacao` |
| `lote` | VARCHAR(50) | |
| `valor_unitario` | DECIMAL(10,2) | |
| `id_usuario_criacao` NOT NULL, `id_usuario_cancelamento` NULL | | auditoria |
| `ativo`, `cancelada_em` | | |

> ⚠️ **`entrada_epi` (a tabela original da migração `000001`) foi inteiramente dropada** numa
> migração posterior e substituída pelo par `entrada_nf` + `entrada_epi_item`. Qualquer coluna
> adicionada a `entrada_epi` por migrações intermediárias (002, 005, 010) morreu junto com o DROP —
> não procure `entrada_epi` no schema atual.

### `entrega_epi` (cabeçalho de uma entrega)

`id`, `tenant_id`, `idfuncionario → funcionario(id)`, `data_entrega DATE`, `assinatura TEXT` (URL
pública no Supabase, nunca base64), `idtroca INT NULL` (auto-referência a outra entrega quando ela
nasce de uma troca em devolução), `cancelada_em`, `ativo`, `token_validacao TEXT NULL` (token de
auditoria `ENT-...`), `id_usuario_entrega NULL → usuarios(id)`,
`id_usuario_entrega_cancelamento NULL → usuarios(id)`.

### `epis_entregues` (uma linha por lote consumido numa entrega)

`id`, `tenant_id`, `id_entrega_cabecalho → entrega_epi(id)`, `id_entrada_item →
entrada_epi_item(id)` (lote de origem exato), `id_epi → epi(id)`, `id_tamanho → tamanho(id)`,
`quantidade`, `ativo`, `deletado_em`.

Uma linha por lote: se um pedido de N unidades precisar consumir 2 lotes diferentes (FEFO), gera 2
linhas em `epis_entregues` para a mesma `entrega_epi`.

> Nota histórica de migração: a FK de `id_entrega_cabecalho` chegou a ser criada apontando (por
> engano) para `entrada_nf(id)` em vez de `entrega_epi(id)`; foi corrigida numa migração seguinte.
> Hoje está correta — só relevante se for mexer em migrações antigas.

### `motivo_devolucao`

`id`, `tenant_id`, `motivo VARCHAR(50)`, `ativo`, `deletado_em`, `gera_descarte BOOL NOT NULL
DEFAULT false`. `UNIQUE(motivo, tenant_id)`.

> A coluna real se chama **`gera_descarte`**, não `eh_descarte` como o `CLAUDE.md` do projeto
> descreve em prosa — o nome de domínio "descarte" é o mesmo, só o identificador da coluna difere.

### `devolucao`

| Coluna | Tipo | Notas |
|---|---|---|
| `id` | SERIAL PK | |
| `tenant_id` | INT | |
| `idepi → epi(id)`, `idfuncionario → funcionario(id)`, `idmotivo → motivo_devolucao(id)`, `idtamanho → tamanho(id)` | | |
| `data_devolucao` | DATE | |
| `quantidadeadevolver` | INT | |
| `idepinovo`, `idtamanhonovo`, `quantidadenova` | NULLABLE | preenchidos só quando há troca |
| `houve_troca` | BOOL DEFAULT false | |
| `observacao` | TEXT NULL | |
| `assinatura_digital` | TEXT | URL Supabase |
| `token_validacao` | TEXT NULL | token de auditoria `DEVO-...` |
| `cancelada_em`, `ativo` | | |
| `id_usuario_cancelamento` | INT NULL → `usuarios(id)` | ⚠️ ver nota abaixo |
| `id_usuario_devolucao_cancelamento` | INT NULL → `usuarios(id)` | ⚠️ ver nota abaixo |

> ⚠️ **Armadilha de nomenclatura confirmada em `database/queries/Devolucao.sql`**: existem duas
> colunas de auditoria de usuário quase idênticas. `id_usuario_cancelamento` é gravado no
> **INSERT** — apesar do nome, na prática registra **quem criou** a devolução, não quem cancelou.
> `id_usuario_devolucao_cancelamento` é o campo setado no **UPDATE** de cancelamento de verdade. É
> fácil interpretar ao contrário só pelo nome — confirme sempre contra a query real antes de
> assumir o significado de uma dessas colunas.

### `planos`

`id`, `nome`, `mensalidade DECIMAL(10,2) NOT NULL`, `limite_funcionarios`, `limite_usuarios`,
`limite_epis` (INT, `NULL` = ilimitado), `status VARCHAR(20) DEFAULT 'Ativo'`, `descricao`,
`criado_em`.

---

## Observações gerais sobre o schema

- **Identificadores sem aspas no `CREATE TABLE`** (`IdDepartamento`, `IdTipoProtecao`, `CA`,
  `Idfornecedor`, `IdTroca`, …) são normalizados para minúsculo pelo Postgres. No banco e nas
  queries do `sqlc` eles aparecem como `iddepartamento`, `idtipoprotecao`, `ca`, `idfornecedor`,
  `idtroca` — não como camelCase.
- **Padrão de soft delete não é uniforme**: a maioria das tabelas de cadastro usa
  `deletado_em`/`ativo`; `entrada_nf`, `entrada_epi_item`, `entrega_epi` e `devolucao` usam
  `cancelada_em`/`ativo`; `fornecedores` usa `cancelado_em`/`ativo` (singular, "cancelado" em vez de
  "cancelada"). Ao escrever uma query nova de listagem, confirme o nome exato da coluna na tabela
  específica em vez de assumir por analogia.
- **Nenhuma tabela tem Row-Level Security.** O isolamento entre empresas é inteiramente feito pela
  aplicação filtrando `tenant_id` em cada query — ver `CLAUDE.md` e [`ARQUITETURA.md`](./ARQUITETURA.md).
