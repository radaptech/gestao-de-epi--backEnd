-- name: CreateEntradaNF :one
INSERT INTO entrada_nf (tenant_id, nota_fiscal_numero, nota_fiscal_serie, data_emissao, id_usuario_criacao, Idfornecedor)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: CreateEntradaEpiItem :exec
INSERT INTO entrada_epi_item (
    tenant_id, entrada_nf_id, id_epi, id_tamanho, quantidade, 
    quantidade_atual, data_fabricacao, data_validade, lote, valor_unitario,id_usuario_criacao
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: ListarEntradas :many
SELECT 
    ei.id, 
    ei.id_epi, 
    e.nome as epi_nome, 
    e.fabricante, 
    e.ca, 
    e.descricao as epi_descricao,
    ei.data_fabricacao, 
    ei.data_validade, 
    e.validade_ca,
    e.IdTipoProtecao, 
    tp.nome as protecao_nome,
    ei.id_tamanho, 
    t.tamanho as tamanho_nome, 
    ei.quantidade, 
    ei.quantidade_atual, 
    nf.data_emissao as data_entrada, -- Vem da NF agora
    ei.lote,
    ei.valor_unitario, 
    nf.Idfornecedor,
    f.nome_fantasia,
    f.razao_social,
    nf.nota_fiscal_numero, 
    nf.nota_fiscal_serie, 
    ei.cancelada_em,
	us.nome as usuario_criacao
FROM entrada_epi_item ei
INNER JOIN entrada_nf nf ON ei.entrada_nf_id = nf.id
INNER JOIN fornecedores f ON nf.Idfornecedor = f.id
INNER JOIN epi e ON ei.id_epi = e.id
INNER JOIN tipo_protecao tp ON e.idtipoprotecao = tp.id
INNER JOIN tamanho t ON ei.id_tamanho = t.id
INNER JOIN usuarios us ON us.id = nf.id_usuario_criacao
WHERE 
    ei.tenant_id = sqlc.arg('tenant_id')
    AND (
        (sqlc.arg('canceladas')::boolean IS FALSE AND ei.cancelada_em IS NULL) OR
        (sqlc.arg('canceladas')::boolean IS TRUE AND ei.cancelada_em IS NOT NULL)
    )
    AND (sqlc.narg('id_epi')::int IS NULL OR ei.id_epi = sqlc.narg('id_epi'))
    AND (sqlc.narg('id_entrada')::int IS NULL OR ei.id = sqlc.narg('id_entrada'))
    AND (sqlc.narg('data_inicio')::date IS NULL OR nf.data_emissao >= sqlc.narg('data_inicio'))
    AND (sqlc.narg('data_fim')::date IS NULL OR nf.data_emissao <= sqlc.narg('data_fim'))
    AND (sqlc.narg('nota_fiscal')::text IS NULL OR nf.nota_fiscal_numero ILIKE '%' || sqlc.narg('nota_fiscal') || '%')
ORDER BY nf.data_emissao ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CancelarEntrada :one
UPDATE entrada_epi_item 
SET 
    cancelada_em = CURRENT_TIMESTAMP, 
    ativo = FALSE,
    id_usuario_cancelamento = $2 
WHERE id = $1 
  AND tenant_id = $3
  AND cancelada_em IS NULL
ReTURNING entrada_nf_id;

-- name: CancelarEntradaNF :execrows
UPDATE entrada_nf 
SET 
    cancelada_em = CURRENT_TIMESTAMP, 
    ativo = FALSE,
    id_usuario_cancelamento = $2 
WHERE id = $1 
  AND tenant_id = $3
  AND cancelada_em IS NULL;


-- name: ContarItensAtivosNF :one
SELECT COUNT(*) 
FROM entrada_epi_item 
WHERE entrada_nf_id = $1 
  AND ativo = TRUE
  AND tenant_id = $2;

-- name: ContarEntradasFiltradas :one
SELECT COUNT(*) 
FROM entrada_epi_item ei
INNER JOIN entrada_nf nf ON ei.entrada_nf_id = nf.id
WHERE 
    ei.tenant_id = sqlc.arg('tenant_id')
    AND (
        (sqlc.arg('canceladas')::boolean IS FALSE AND ei.cancelada_em IS NULL) OR
        (sqlc.arg('canceladas')::boolean IS TRUE AND ei.cancelada_em IS NOT NULL)
    )
    AND (sqlc.narg('id_epi')::int IS NULL OR ei.id_epi = sqlc.narg('id_epi'))
    AND (sqlc.narg('data_inicio')::date IS NULL OR nf.data_emissao >= sqlc.narg('data_inicio'))
    AND (sqlc.narg('nota_fiscal')::text IS NULL OR nf.nota_fiscal_numero ILIKE '%' || sqlc.narg('nota_fiscal') || '%');

-- name: EntradaDashbord :many
SELECT
    ei.id, ei.id_epi, ei.id_tamanho, ei.quantidade_atual, ei.valor_unitario, ei.quantidade,
    nf.data_emissao as data_entrada, ei.lote
FROM entrada_epi_item ei
INNER JOIN entrada_nf nf ON ei.entrada_nf_id = nf.id
WHERE ei.tenant_id = $1 AND ei.ativo = TRUE
ORDER BY nf.data_emissao DESC;

-- name: EntradaEpiEstoque :many
SELECT
    ei.id, ei.lote, ei.quantidade as quantidade_inicial, ei.quantidade_atual,
    ei.valor_unitario, ei.data_validade,
    ei.id_tamanho, t.tamanho as tamanho_nome,
    ei.id_epi, e.nome as epi_nome, e.fabricante, e.ca, e.descricao, e.validade_ca, e.alerta_minimo,
    e.idtipoprotecao, tp.nome as protecao_nome,
    nf.data_emissao as data_entrada
FROM entrada_epi_item ei
INNER JOIN entrada_nf nf ON ei.entrada_nf_id = nf.id
INNER JOIN tamanho t ON ei.id_tamanho = t.id
INNER JOIN epi e ON ei.id_epi = e.id
INNER JOIN tipo_protecao tp ON e.idtipoprotecao = tp.id
WHERE ei.tenant_id = $1
  AND ei.ativo = TRUE
ORDER BY nf.data_emissao ASC;