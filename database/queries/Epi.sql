-- name: AddEpi :one
INSERT INTO epi (tenant_id, nome, fabricante, CA, descricao, validade_CA, IdTipoProtecao, alerta_minimo) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: AddEpiTamanho :exec
INSERT INTO tamanhos_epis (tenant_id, IdEpi, IdTamanho, ativo, deletado_em) 
VALUES ($1, $2, $3, TRUE, NULL)
ON CONFLICT (IdEpi, IdTamanho, tenant_id) 
DO UPDATE SET ativo = TRUE, deletado_em = NULL;

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
SET ativo = FALSE, deletado_em = current_date
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE;

-- name: DeletarTamanhosPorEpi :execrows

DELETE FROM tamanhos_epis 
WHERE IdEpi = $1 AND tenant_id = $2;

-- name: UpdateEpiCampo :execrows
UPDATE epi 
SET 
    nome = COALESCE(sqlc.narg('nome'), nome),
    fabricante = COALESCE(sqlc.narg('fabricante'), fabricante),
    CA = COALESCE(sqlc.narg('ca'), CA),
    descricao = COALESCE(sqlc.narg('descricao'), descricao),
    validade_CA = COALESCE(sqlc.narg('validade_ca'), validade_CA),
    IdTipoProtecao = COALESCE(sqlc.narg('id_tipo_protecao')::int, IdTipoProtecao),
    alerta_minimo = COALESCE(sqlc.narg('alerta_minimo')::int, alerta_minimo)  
    WHERE id = sqlc.arg('id') 
  AND tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Obrigatório para update
  AND ativo = TRUE;


-- name: BuscaEpiDashbord :many
SELECT id, nome, alerta_minimo
FROM epi
WHERE tenant_id = $1 -- SEGURANÇA
  AND ativo = TRUE
ORDER BY nome;


-- name: BuscaTodosItensEntreguesDoTenant :many
SELECT
    ee.id,
    ee.id_entrega_cabecalho, -- 🔑 CHAVE MESTRA: Liga o item à entrega correta no Go!
    ee.quantidade,
    ee.id_tamanho,
    t.tamanho as tamanho_nome,
    ee.id_epi,
    ep.nome as epi_nome,
    ep.fabricante,
    ep.CA,
    ep.descricao,
    ep.validade_CA,
    ep.alerta_minimo,
    ep.IdTipoProtecao,
    tp.nome as tipo_protecao_nome
FROM epis_entregues ee
INNER JOIN entrega_epi e ON e.id = ee.id_entrega_cabecalho
INNER JOIN epi ep ON ep.id = ee.id_epi
INNER JOIN tamanho t ON t.id = ee.id_tamanho
INNER JOIN tipo_protecao tp ON tp.id = ep.IdTipoProtecao
WHERE e.tenant_id = $1;


-- name: BuscaItensEntreguesPorFuncionario :many
WITH Entregas AS (
    -- 1. Agrupa e soma todas as quantidades que o funcionário já recebeu por EPI e Tamanho
    SELECT 
        ee.id_epi, 
        ee.id_tamanho, 
        SUM(ee.quantidade)::int as total_entregue
    FROM epis_entregues ee
    INNER JOIN entrega_epi e ON e.id = ee.id_entrega_cabecalho
    WHERE e.tenant_id = $1 
      AND e.IdFuncionario = $2
      AND e.cancelada_em IS NULL
    GROUP BY ee.id_epi, ee.id_tamanho
),
Devolucoes AS (
    -- 2. Agrupa e soma tudo que ele já devolveu desse EPI e Tamanho
    SELECT 
      d.idepi, 
        d.idtamanho, 
        SUM(d.quantidadeadevolver)::int as total_devolvido
    FROM devolucao d
    WHERE d.tenant_id = $1 
      AND d.idfuncionario = $2
    GROUP BY d.idepi, d.idtamanho
)
-- 3. Junta as tabelas, traz os dados completos do EPI e filtra o saldo
SELECT 
    Entregas.id_epi,
    ep.nome as epi_nome,
    ep.fabricante,
    ep.ca,
    ep.descricao,
    ep.validade_ca,
    ep.alerta_minimo,
    ep.idtipoprotecao,
    tp.nome as tipo_protecao_nome,
    Entregas.id_tamanho,
    t.tamanho as tamanho_nome,
    (Entregas.total_entregue - COALESCE(Devolucoes.total_devolvido, 0))::int AS saldo_atual
FROM Entregas
LEFT JOIN Devolucoes 
    ON Entregas.id_epi = Devolucoes.idepi 
   AND Entregas.id_tamanho = Devolucoes.idtamanho
INNER JOIN epi ep ON ep.id = Entregas.id_epi
INNER JOIN tamanho t ON t.id = Entregas.id_tamanho
INNER JOIN tipo_protecao tp ON tp.id = ep.IdTipoProtecao
WHERE (Entregas.total_entregue - COALESCE(Devolucoes.total_devolvido, 0)) > 0;