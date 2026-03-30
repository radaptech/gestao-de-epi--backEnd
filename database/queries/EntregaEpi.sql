-- name: AddEntregaEpi :one
INSERT INTO entrega_epi (
    tenant_id, -- Novo campo
    IdFuncionario, data_entrega, assinatura, IdTroca, token_validacao, id_usuario_entrega
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: AddItemEntregue :one
INSERT INTO epis_entregues (
    tenant_id, -- Novo campo (redundante mas necessário para segurança/performance)
    IdEntrega, IdEntrada, IdEpi, IdTamanho, quantidade
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, IdEntrega;

-- name: CancelaItemEntregue :many
UPDATE epis_entregues
SET ativo = FALSE, deletado_em = NOW()
WHERE IdEntrega = $1 
  AND tenant_id = $2 -- SEGURANÇA
RETURNING IdEntrada, quantidade;

-- name: ListarEntregas :many
SELECT 
    ee.id as entrega_id, ee.data_entrega, ee.assinatura, ee.token_validacao, ee.id_usuario_entrega,   
    f.id as func_id, f.nome as func_nome, f.matricula,
    d.id as dep_id, d.nome as dep_nome,
    ff.id as funcao_id, ff.nome as funcao_nome,
    e.id as epi_id, e.nome as epi_nome, e.fabricante, e.CA, e.descricao as epi_desc, e.validade_CA,
    tp.id as tp_id, tp.nome as tp_nome,
    t.id as tam_id, t.tamanho as tam_nome,
    i.quantidade,
    COUNT(*) OVER() as total_geral
FROM entrega_epi ee
INNER JOIN funcionario f ON ee.IdFuncionario = f.id
INNER JOIN departamento d ON f.IdDepartamento = d.id
INNER JOIN funcao ff ON f.IdFuncao = ff.id
INNER JOIN epis_entregues i ON i.IdEntrega = ee.id
INNER JOIN epi e ON i.IdEpi = e.id
INNER JOIN tipo_protecao tp ON e.IdTipoProtecao = tp.id
INNER JOIN tamanho t ON i.IdTamanho = t.id
WHERE 
    ee.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Filtro Principal
    AND (
        (sqlc.arg('canceladas')::boolean IS FALSE AND ee.cancelada_em IS NULL) OR
        (sqlc.arg('canceladas')::boolean IS TRUE AND ee.cancelada_em IS NOT NULL)
    )
    AND (sqlc.narg('id_entrega')::int IS NULL OR ee.id = sqlc.narg('id_entrega'))
    AND (sqlc.narg('id_funcionario')::int IS NULL OR ee.IdFuncionario = sqlc.narg('id_funcionario'))
    AND (sqlc.narg('data_inicio')::date IS NULL OR ee.data_entrega >= sqlc.narg('data_inicio'))
    AND (sqlc.narg('data_fim')::date IS NULL OR ee.data_entrega <= sqlc.narg('data_fim'))
ORDER BY ee.data_entrega DESC
LIMIT $1 OFFSET $2;

-- name: CancelarEntrega :one
UPDATE entrega_epi
SET cancelada_em =current_date,
    ativo = FALSE,
    id_usuario_entrega_cancelamento = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND cancelada_em IS NULL
RETURNING id;

-- name: BuscarTodosItensEntrega :many
SELECT 
    i.IdEntrega as entrega_id, i.id as item_id, i.quantidade,
    e.id as epi_id, e.nome as epi_nome, e.fabricante, e.CA, e.descricao as epi_desc, e.validade_CA,
    tp.id as tp_id, tp.nome as tp_nome,
    t.id as tam_id, t.tamanho as tam_nome
FROM epis_entregues i
INNER JOIN epi e ON i.IdEpi = e.id
INNER JOIN tipo_protecao tp ON e.IdTipoProtecao = tp.id
INNER JOIN tamanho t ON i.IdTamanho = t.id
WHERE 
    i.tenant_id = sqlc.arg('tenant_id') 
    AND i.IdEntrega = sqlc.arg('id_entrega') -- FALTOU ISSO
    AND i.ativo = TRUE;

-- name: ListarItensEntregueCancelados :many
SELECT quantidade, IdEntrada
FROM epis_entregues
WHERE IdEntrega = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = FALSE 
  AND deletado_em IS NOT NULL;

-- name: CancelaEntregaPorIdTroca :one
UPDATE entrega_epi
SET cancelada_em =current_date,
    ativo = FALSE,
    id_usuario_entrega_cancelamento = $2
WHERE IdTroca = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND cancelada_em IS NULL
RETURNING id;

-- name: ListarHistoricoEntregasPorMatricula :many
SELECT 
    emp.razao_social as razao_social,
    f.id as func_id, f.nome as func_nome, f.matricula,
    d.id as dep_id, d.nome as dep_nome,
    ff.id as funcao_id, ff.nome as funcao_nome,
    ee.data_entrega,i.quantidade,e.CA,e.nome AS epi_nome, e.descricao,
    i.IdTamanho,t.tamanho,ee.assinatura
FROM entrega_epi ee
INNER JOIN empresas emp ON ee.tenant_id = emp.id
INNER JOIN funcionario f ON ee.IdFuncionario = f.id
INNER JOIN departamento d ON f.IdDepartamento = d.id
INNER JOIN funcao ff ON f.IdFuncao = ff.id
INNER JOIN epis_entregues i ON i.IdEntrega = ee.id
INNER JOIN epi e ON e.id = i.IdEpi
INNER JOIN tamanho t ON t.id = i.IdTamanho
where f.matricula = $1
and ee.tenant_id = $2
and ee.ativo = TRUE
ORDER BY ee.data_entrega DESC,ee.id DESC;

-- name: EntregaDashbord :many
SELECT
    id, Idfuncionario,data_entrega,
    assinatura,token_validacao
FROM entrega_epi
WHERE tenant_id = $1 and ativo = TRUE
ORDER BY data_entrega DESC;


-- name: EntregaItensDashbord :many
SELECT
    id, IdEntrega, IdEpi, IdTamanho, quantidade
FROM epis_entregues
WHERE tenant_id = $1 and ativo = TRUE
ORDER BY id DESC;