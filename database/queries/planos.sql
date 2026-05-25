-- name: AddPlano :one
INSERT INTO planos (nome, mensalidade, limite_funcionarios, limite_usuarios, limite_epis, status, descricao)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id;

-- name: BuscaPlanos :many
SELECT id, nome, mensalidade, limite_funcionarios, limite_usuarios, limite_epis, status, descricao
FROM planos
WHERE status = 'Ativo'
ORDER BY nome ASC;