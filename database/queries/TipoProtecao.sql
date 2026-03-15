-- name: AddProtecao :one
INSERT INTO tipo_protecao (tenant_id, nome) 
VALUES ($1, $2)
RETURNING *;

-- name: BuscarProtecao :one
SELECT id, nome 
FROM tipo_protecao 
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE 
LIMIT 1;

-- name: BuscarTodasProtecoes :many
SELECT id, nome 
FROM tipo_protecao 
WHERE tenant_id = $1 -- SEGURANÇA: Lista apenas do cliente logado
  AND ativo = TRUE
ORDER BY nome ASC;

-- name: DeletarProtecao :execrows
UPDATE tipo_protecao
SET ativo = FALSE,
    deletado_em = NOW()
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE;

-- name: UpdateProtecao :execrows
UPDATE tipo_protecao
SET nome = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND ativo = TRUE;