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


-- name: RecuperaLogin :one
select id from usuarios 
where email = $1 and tenant_id = $2 limit 1;

-- name: SalvarTokenRecuperacao :execrows
update usuarios
set token_recuperacao_senha = $1, token_expiracao = $2
where id = $3 and tenant_id = $4;   

-- name: UpdateSenha :execrows
update usuarios
set senha_hash = $1,
    token_recuperacao_senha = null,
    token_expiracao = null
where token_recuperacao_senha = $2 and tenant_id = $3 and token_expiracao > now();


-- name: AtualizarUltimoAcesso :execrows
UPDATE usuarios
SET ultimo_acesso = now() AT TIME ZONE 'America/Sao_Paulo'
WHERE id = $1 AND tenant_id = $2;


-- name: MostrarUsuariosPainel :many
select tenant_id, nome, u.email, e.nome_fantasia as empresa,role as tipo,ativo, ultimo_acesso
from usuarios u
inner join empresas e on u.tenant_id = e.id;


-- name: EditarUsuario :exec
UPDATE usuarios
SET 
    nome = @nome,
    email = @email,
    role = @role
WHERE id = @id;