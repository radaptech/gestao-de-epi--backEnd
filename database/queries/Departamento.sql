-- name: CriaDepartamento :one
INSERT INTO departamento (tenant_id, nome) 
VALUES ($1, $2)
ON CONFLICT (tenant_id, nome) DO NOTHING
RETURNING *;

-- name: BuscarDepartamento :one
SELECT id, nome 
FROM departamento 
WHERE id = $1 
  AND tenant_id = $2 
  AND ativo = TRUE 
LIMIT 1;


-- name: BuscarTodosDepartamentos :many
SELECT id, nome as departamento,
COUNT(*) OVER() AS total_geral
FROM departamento 
WHERE tenant_id = sqlc.arg('tenant_id') 
  AND (sqlc.narg('nome')::text IS NULL OR nome ILIKE '%' || sqlc.narg('nome') || '%')
  AND (sqlc.narg('id')::int IS NULL OR id = sqlc.narg('id')::int)
  AND (

    (sqlc.arg('cancelados')::boolean IS FALSE AND deletado_em IS NULL) OR
    (sqlc.arg('cancelados')::boolean IS TRUE AND deletado_em IS NOT NULL)
  )
ORDER BY nome ASC
LIMIT $1 OFFSET $2;

-- name: DeletarDepartamento :execrows
UPDATE departamento
SET ativo = FALSE,
    deletado_em = current_date
WHERE id = $1 
  AND tenant_id = $2 
  AND ativo = TRUE;

-- name: UpdateDepartamento :execrows
UPDATE departamento
SET nome = $2
WHERE id = $1 
  AND tenant_id = $3 
  AND ativo = TRUE;