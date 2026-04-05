-- name: AddEntregaEpi :one
INSERT INTO entrega_epi (
    tenant_id, 
    IdFuncionario, data_entrega, assinatura, IdTroca, token_validacao, id_usuario_entrega
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: AddItemEntregue :one
INSERT INTO epis_entregues (
    tenant_id, 
    id_entrega_cabecalho, id_entrada_item, id_epi, id_tamanho, quantidade
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, id_entrega_cabecalho;

-- name: CancelaItemEntregue :many
UPDATE epis_entregues
SET ativo = FALSE, deletado_em = NOW()
WHERE id_entrega_cabecalho = $1 
  AND tenant_id = $2
RETURNING id_entrada_item, quantidade;

-- name: ListarEntregas :many
SELECT 
    ee.id as entrega_id, ee.data_entrega, ee.assinatura, ee.token_validacao, ee.id_usuario_entrega,    
    f.id as func_id, f.nome as func_nome, f.matricula,
    d.id as dep_id, d.nome as dep_nome,
    ff.id as funcao_id, ff.nome as funcao_nome,
    COUNT(*) OVER() as total_geral
FROM entrega_epi ee
INNER JOIN funcionario f ON ee.idfuncionario = f.id
INNER JOIN departamento d ON f.iddepartamento = d.id
INNER JOIN funcao ff ON f.idfuncao = ff.id
WHERE 
    ee.tenant_id = sqlc.arg('tenant_id')
    AND (
        (sqlc.arg('canceladas')::boolean IS FALSE AND ee.cancelada_em IS NULL) OR
        (sqlc.arg('canceladas')::boolean IS TRUE AND ee.cancelada_em IS NOT NULL)
    )
    AND (sqlc.narg('id_entrega')::int IS NULL OR ee.id = sqlc.narg('id_entrega'))
    AND (sqlc.narg('idfuncionario')::int IS NULL OR ee.idfuncionario = sqlc.narg('idfuncionario'))
ORDER BY ee.data_entrega DESC
LIMIT $1 OFFSET $2;

-- name: CancelarEntrega :one
UPDATE entrega_epi
SET cancelada_em = CURRENT_DATE,
    ativo = FALSE,
    id_usuario_entrega_cancelamento = $2
WHERE id = $1 
  AND tenant_id = $3
  AND cancelada_em IS NULL
RETURNING id;

-- name: BuscarTodosItensEntrega :many
SELECT 
    i.id_entrega_cabecalho as entrega_id, i.id as item_id, i.quantidade,
    e.id as epi_id, e.nome as epi_nome, e.fabricante, e.ca, e.descricao as epi_desc, e.validade_ca,
    tp.id as tp_id, tp.nome as tp_nome,
    t.id as tam_id, t.tamanho as tam_nome
FROM epis_entregues i
INNER JOIN epi e ON i.id_epi = e.id
INNER JOIN tipo_protecao tp ON e.idtipoprotecao = tp.id
INNER JOIN tamanho t ON i.id_tamanho = t.id
WHERE 
    i.tenant_id = sqlc.arg('tenant_id') 
    AND (sqlc.arg('id_entrega')::int = 0 OR i.id_entrega_cabecalho = sqlc.arg('id_entrega'))
    AND i.ativo = TRUE;

-- name: ListarItensEntregueCancelados :many
SELECT quantidade, id_entrada_item
FROM epis_entregues
WHERE id_entrega_cabecalho = $1 
  AND tenant_id = $2
  AND ativo = FALSE 
  AND deletado_em IS NOT NULL;

-- name: CancelaEntregaPorIdTroca :one
UPDATE entrega_epi
SET cancelada_em = CURRENT_DATE,
    ativo = FALSE,
    id_usuario_entrega_cancelamento = $2
WHERE IdTroca = $1 
  AND tenant_id = $3
  AND cancelada_em IS NULL
RETURNING id;

-- name: ListarHistoricoEntregasPorMatricula :many
SELECT 
    emp.razao_social as razao_social,
    f.id as func_id, f.nome as func_nome, f.matricula,
    d.id as dep_id, d.nome as dep_nome,
    ff.id as funcao_id, ff.nome as funcao_nome,
    ee.data_entrega, i.quantidade, e.ca, e.nome AS epi_nome, e.descricao,
    i.id_tamanho, t.tamanho, ee.assinatura
FROM entrega_epi ee
INNER JOIN empresas emp ON ee.tenant_id = emp.id
INNER JOIN funcionario f ON ee.id_funcionario = f.id
INNER JOIN departamento d ON f.id_departamento = d.id
INNER JOIN funcao ff ON f.id_funcao = ff.id
INNER JOIN epis_entregues i ON i.id_entrega_cabecalho = ee.id
INNER JOIN epi e ON e.id = i.id_epi
INNER JOIN tamanho t ON t.id = i.id_tamanho
WHERE f.matricula = $1
AND ee.tenant_id = $2
AND ee.ativo = TRUE
ORDER BY ee.data_entrega DESC, ee.id DESC;

-- name: EntregaDashbord :many
SELECT
    id, IdFuncionario, data_entrega,
    assinatura, token_validacao
FROM entrega_epi
WHERE tenant_id = $1 AND ativo = TRUE
ORDER BY data_entrega DESC;

-- name: EntregaItensDashbord :many
SELECT
    id, id_entrega_cabecalho, id_epi, id_tamanho, quantidade
FROM epis_entregues
WHERE tenant_id = $1 AND ativo = TRUE
ORDER BY id DESC;

-- name: BuscaTodasEntregasDoTenant :many
SELECT 
    id, 
    IdFuncionario,
    data_entrega, 
    assinatura
FROM entrega_epi
WHERE tenant_id = $1 AND ativo = TRUE
ORDER BY data_entrega DESC;