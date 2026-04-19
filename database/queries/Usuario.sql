-- name: CreateUser :exec 
INSERT INTO usuarios (tenant_id, nome, email, senha_hash, role) 
VALUES ($1, $2, $3, $4, $5);

-- name: BuscarPorIdUsuario :one
SELECT id, nome, email, ativo, role
FROM usuarios
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE
LIMIT 1;

-- name: BuscarTodosUsuarios :many
SELECT id, nome, email, role as cargo
FROM usuarios
WHERE tenant_id = $1 -- SEGURANÇA: Lista apenas usuários desta empresa
  AND ativo = TRUE;

-- name: DeletarUsuario :execrows
UPDATE usuarios
SET ativo = FALSE
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE; 

-- name: BuscarUsuarioPorEmail :one
-- Atenção: Se o email puder se repetir entre empresas, o tenant_id é OBRIGATÓRIO aqui.
SELECT id, nome, email, senha_hash, tenant_id, role
FROM usuarios
WHERE email = $1 
  AND tenant_id = $2 
  AND ativo = TRUE
LIMIT 1;