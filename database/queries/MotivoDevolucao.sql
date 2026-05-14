-- name: AddMotivoDevolucao :one
INSERT INTO motivo_devolucao (tenant_id, motivo, gera_descarte) 
VALUES ($1, $2, $3)
returning id, motivo;

-- name: BuscaMotivoDevolucao :one
SELECT id, motivo, gera_descarte
FROM motivo_devolucao 
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE 
LIMIT 1;

-- name: BuscaTodosMotivosDevolucao :many
SELECT id, motivo, gera_descarte
FROM motivo_devolucao 
WHERE tenant_id = $1 -- SEGURANÇA: Lista apenas os motivos desta empresa
  AND ativo = TRUE
ORDER BY motivo ASC;

-- name: DeleteMotivoDevolucao :execrows
UPDATE motivo_devolucao
SET ativo = FALSE,
    deletado_em = current_date
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE;

-- name: EhDescarte :one
SELECT gera_descarte
FROM motivo_devolucao
WHERE id = $1
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE
LIMIT 1;