-- name: AddFuncao :exec
INSERT INTO funcao (tenant_id, nome, IdDepartamento) 
VALUES ($1, $2, $3);

-- name: BuscarFuncao :one
SELECT 
    f.id, 
    f.nome, 
    f.IdDepartamento, 
    d.nome as departamento_nome
FROM funcao f
INNER JOIN departamento d ON f.IdDepartamento = d.id
WHERE f.id = $1 
  AND f.tenant_id = $2 -- SEGURANÇA
  AND f.ativo = TRUE;

-- name: BuscarTodasFuncoes :many
SELECT 
    f.id, 
    f.nome, 
    f.IdDepartamento, 
    d.nome as departamento_nome,
    COUNT(*) OVER() AS total_geral
FROM funcao f
INNER JOIN departamento d ON f.IdDepartamento = d.id
WHERE f.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA
  AND (sqlc.narg('nome')::text IS NULL OR f.nome ILIKE '%' || sqlc.narg('nome') || '%')
  AND (sqlc.narg('id')::int IS NULL OR f.id = sqlc.narg('id')::int)
  AND (
    (sqlc.arg('cancelados')::boolean IS FALSE AND f.deletado_em IS NULL) OR
    (sqlc.arg('cancelados')::boolean IS TRUE AND f.deletado_em IS NOT NULL)
  )
  AND (sqlc.narg('nome_departamento')::text IS NULL OR d.nome ILIKE '%' || sqlc.narg('nome_departamento') || '%')
ORDER BY f.nome ASC
LIMIT $1 OFFSET $2;

-- name: DeletarFuncao :execrows
UPDATE funcao
SET ativo = FALSE,
    deletado_em = current_date
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE;

-- name: UpdateFuncao :execrows
UPDATE funcao
SET nome = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND ativo = TRUE;