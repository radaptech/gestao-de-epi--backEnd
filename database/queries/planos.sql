-- name: AddPlano :one
INSERT INTO planos (nome, mensalidade, limite_funcionarios, limite_usuarios, limite_epis, status, descricao)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id;

-- name: BuscaPlanos :many
SELECT id, nome, mensalidade, limite_funcionarios, limite_usuarios, limite_epis, status, descricao
FROM planos
WHERE status = 'Ativo'
ORDER BY nome ASC;

-- name: AtualizarPlano :exec
UPDATE planos
SET
  nome = COALESCE(sqlc.narg('nome'), nome),
  mensalidade = COALESCE(sqlc.narg('mensalidade'), mensalidade),
  descricao = COALESCE(sqlc.narg('descricao'), descricao),
  limite_funcionarios = COALESCE(sqlc.narg('limite_funcionarios'), limite_funcionarios),
  limite_usuarios = COALESCE(sqlc.narg('limite_usuarios'), limite_usuarios),
  limite_epis = COALESCE(sqlc.narg('limite_epis'), limite_epis),
  status = COALESCE(sqlc.narg('status'), status)
WHERE id = sqlc.arg('id');


-- name: AtualizarStatusPlano :exec
UPDATE planos
SET status = sqlc.arg('status')
WHERE id = sqlc.arg('id');

-- name: BuscarPlanoPorNome :one
select id, nome from planos 
where nome = $1 and status = 'Ativo';