-- name: ListarLotesParaConsumo :many
-- O PostgreSQL usa FOR UPDATE para travar apenas as linhas desse cliente específico.
SELECT id, quantidadeAtual, data_validade, valor_unitario 
FROM entrada_epi 
WHERE tenant_id = $1 -- SEGURANÇA: Só busca lotes da empresa logada
  AND IdEpi = $2 
  AND IdTamanho = $3 
  AND quantidadeAtual > 0 
  AND data_validade >= CURRENT_DATE
  AND ativo = TRUE
ORDER BY data_validade ASC
FOR UPDATE;

-- name: AbaterEstoqueLote :execrows
UPDATE entrada_epi 
SET quantidadeAtual = quantidadeAtual - $1 
WHERE id = $2 
  AND tenant_id = $3 -- SEGURANÇA: Garante que o lote pertence à empresa antes de subtrair
  AND ativo = TRUE
  AND quantidadeAtual >= $1;

-- name: ReporEstoqueLote :execrows
UPDATE entrada_epi 
SET quantidadeAtual = quantidadeAtual + $1 
WHERE id = $2 
  AND tenant_id = $3 -- SEGURANÇA
  AND ativo = TRUE;

-- name: RegistrarItemEntrega :exec
INSERT INTO epis_entregues (
    tenant_id, -- Novo campo
    IdEpi, IdTamanho, quantidade, IdEntrega, IdEntrada
) 
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DevolverItemAoEstoque :exec
UPDATE entrada_epi
SET quantidadeAtual = entrada_epi.quantidadeAtual + $4 -- Quantidade é o $4 agora
WHERE id = (
    SELECT ee.id
    FROM entrada_epi ee
    WHERE ee.tenant_id = $1 -- SEGURANÇA NA SUBQUERY
      AND ee.IdEpi = $2 
      AND ee.IdTamanho = $3
      AND ee.ativo = TRUE -- Boa prática garantir que não devolve para lote inativo
    ORDER BY ee.data_entrada DESC
    LIMIT 1
)
AND tenant_id = $1; -- SEGURANÇA NO UPDATE


-- name: ListarEstoqueAtual :many
SELECT 
    e.IdEpi, 
    p.nome AS nome_epi,
    SUM(e.quantidadeAtual) AS quantidade_total,
    COUNT(*) OVER() AS total_geral
FROM entrada_epi e
inner JOIN epi p ON e.IdEpi = p.id
WHERE e.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Só busca estoque da empresa logada
  AND e.ativo = TRUE
GROUP BY e.IdEpi, p.nome
LIMIT $1 OFFSET $2;


-- name: ListarSaldoEstoque :many
SELECT 
    e.IdEpi, 
    p.nome AS nome_epi,
    SUM(e.quantidadeAtual)::int AS quantidade_atual,
    SUM(e.valor_unitario * e.quantidadeAtual )::float AS saldo_atual,
    COUNT(*) OVER() AS total_geral
    from entrada_epi e
inner JOIN epi p ON e.IdEpi = p.id
WHERE e.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Só busca estoque da empresa logada
  AND e.ativo = TRUE
  AND p.fabricante = sqlc.narg('fabricante') OR sqlc.narg('fabricante') IS NULL -- Filtro adicional para fabricante
GROUP BY e.IdEpi, p.nome
LIMIT $1 OFFSET $2;