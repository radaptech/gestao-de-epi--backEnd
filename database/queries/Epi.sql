-- name: AddEpi :one
INSERT INTO epi (tenant_id, nome, fabricante, CA, descricao, validade_CA, IdTipoProtecao, alerta_minimo) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: AddEpiTamanho :exec
INSERT INTO tamanhos_epis (tenant_id, IdEpi, IdTamanho) 
VALUES ($1, $2, $3);

-- name: BuscarEpi :one
SELECT 
    e.id, e.nome, e.fabricante, e.CA, e.descricao,
    e.validade_CA, e.alerta_minimo, e.IdTipoProtecao, 
    tp.nome as tipo_protecao_nome
FROM epi e
INNER JOIN tipo_protecao tp ON e.IdTipoProtecao = tp.id
WHERE e.id = $1 
  AND e.tenant_id = $2 -- SEGURANÇA
  AND e.ativo = TRUE;

-- name: BuscarTamanhosPorEpi :many
SELECT t.id, t.tamanho
FROM tamanho t
INNER JOIN tamanhos_epis te ON t.id = te.IdTamanho
WHERE te.IdEpi = $1 
  AND te.tenant_id = $2 -- SEGURANÇA: Garante que a relação pertence à empresa
  AND te.ativo = TRUE;

-- name: BuscarTodosEpisPaginado :many
SELECT 
    e.id, e.nome, e.fabricante, e.CA, e.descricao,
    e.validade_CA, e.alerta_minimo, e.IdTipoProtecao, 
    tp.nome as tipo_protecao_nome,
    COUNT(*) OVER() as total_geral
FROM epi e
INNER JOIN tipo_protecao tp ON e.IdTipoProtecao = tp.id
WHERE e.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Filtro de Tenant
  AND (sqlc.narg('nome')::text IS NULL OR e.nome ILIKE '%' || sqlc.narg('nome') || '%')
  AND (sqlc.narg('ca')::text IS NULL OR e.CA ILIKE '%' || sqlc.narg('ca') || '%')
  AND (sqlc.narg('id')::int IS NULL OR e.id = sqlc.narg('id')::int)
  AND (
    (sqlc.arg('cancelados')::boolean IS FALSE AND e.deletado_em IS NULL) OR
    (sqlc.arg('cancelados')::boolean IS TRUE AND e.deletado_em IS NOT NULL)
  )
  AND (sqlc.narg('fabricante')::text IS NULL OR e.fabricante ILIKE '%' || sqlc.narg('fabricante') || '%')
ORDER BY e.id
LIMIT $1 OFFSET $2;

-- name: BuscarTodosTamanhosAgrupados :many
SELECT te.IdEpi, t.id, t.tamanho
FROM tamanho t
INNER JOIN tamanhos_epis te ON t.id = te.IdTamanho
WHERE te.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA
  AND te.ativo = TRUE;

-- name: DeletarEpi :execrows
UPDATE epi 
SET ativo = FALSE, deletado_em = NOW() 
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE;

-- name: DeletarTamanhosPorEpi :execrows
UPDATE tamanhos_epis 
SET ativo = FALSE, deletado_em = NOW() 
WHERE IdEpi = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE;

-- name: UpdateEpiCampo :execrows
UPDATE epi 
SET 
    nome = COALESCE(sqlc.narg('nome'), nome),
    fabricante = COALESCE(sqlc.narg('fabricante'), fabricante),
    CA = COALESCE(sqlc.narg('ca'), CA),
    descricao = COALESCE(sqlc.narg('descricao'), descricao),
    validade_CA = COALESCE(sqlc.narg('validade_ca'), validade_CA)
WHERE id = sqlc.arg('id') 
  AND tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Obrigatório para update
  AND ativo = TRUE;