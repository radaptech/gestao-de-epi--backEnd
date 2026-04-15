-- name: ListarLotesParaConsumo :many
-- O PostgreSQL usa FOR UPDATE para travar apenas as linhas desse cliente específico.
SELECT id, quantidade_atual, data_validade, valor_unitario 
FROM entrada_epi_item 
WHERE tenant_id = $1 
  AND id_epi = $2 
  AND id_tamanho = $3 
  AND quantidade_atual > 0 
  AND data_validade >= CURRENT_DATE
  AND ativo = TRUE
ORDER BY data_validade ASC
FOR UPDATE;

-- name: AbaterEstoqueLote :execrows
UPDATE entrada_epi_item 
SET quantidade_atual = quantidade_atual - $1 
WHERE id = $2 
  AND tenant_id = $3 
  AND ativo = TRUE
  AND quantidade_atual >= $1;

-- name: ReporEstoqueLote :execrows
UPDATE entrada_epi_item 
SET quantidade_atual = quantidade_atual + $1 
WHERE id = $2 
  AND tenant_id = $3 
  AND ativo = TRUE;

-- name: RegistrarItemEntrega :exec
INSERT INTO epis_entregues (
    tenant_id, 
    id_epi, 
    id_tamanho, 
    quantidade, 
    id_entrega_cabecalho, 
    id_entrada_item
) 
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DevolverItemAoEstoque :exec
UPDATE entrada_epi_item
SET quantidade_atual = entrada_epi_item.quantidade_atual + $4 -- 👈 Explicitamos a tabela aqui
WHERE id = (
    SELECT eei.id
    FROM entrada_epi_item eei
    WHERE eei.tenant_id = $1 
      AND eei.id_epi = $2 
      AND eei.id_tamanho = $3
      AND eei.ativo = TRUE
    ORDER BY eei.id DESC 
    LIMIT 1
)
AND tenant_id = $1;

-- name: ListarEstoqueAtual :many
SELECT 
    e.id_epi, 
    p.nome AS nome_epi,
    SUM(e.quantidade_atual)::bigint AS quantidade_total,
    COUNT(*) OVER() AS total_geral
FROM entrada_epi_item e
INNER JOIN epi p ON e.id_epi = p.id
WHERE e.tenant_id = sqlc.arg('tenant_id')
  AND e.ativo = TRUE
GROUP BY e.id_epi, p.nome
LIMIT $1 OFFSET $2;

-- name: ListarSaldoEstoque :many
SELECT 
    e.id_epi, 
    p.nome AS nome_epi,
    SUM(e.quantidade_atual)::int AS quantidade_atual,
    SUM(e.valor_unitario * e.quantidade_atual)::float AS saldo_atual,
    COUNT(*) OVER() AS total_geral
FROM entrada_epi_item e
INNER JOIN epi p ON e.id_epi = p.id
WHERE e.tenant_id = sqlc.arg('tenant_id')
  AND e.ativo = TRUE
  AND (p.fabricante = sqlc.narg('fabricante') OR sqlc.narg('fabricante') IS NULL)
GROUP BY e.id_epi, p.nome
LIMIT $1 OFFSET $2;


