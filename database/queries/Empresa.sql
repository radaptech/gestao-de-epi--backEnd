-- name: GetTenantBySubdomain :one
SELECT id, nome_fantasia 
FROM empresas 
WHERE subdominio = $1 AND status = 'Ativa';

-- name: CriarEmpresa :exec
INSERT INTO empresas (
    nome_fantasia,
    razao_social,
    cnpj,
    responsavel,
    email,
    telefone,
    plano_id,
    status,
    mensalidade,
    vencimento,
    observacoes,
    subdominio
) VALUES (
    sqlc.arg('nome_fantasia'),
    sqlc.arg('razao_social'),
    sqlc.arg('cnpj'),
    sqlc.arg('responsavel'),
    sqlc.arg('email'),
    sqlc.arg('telefone'),
    sqlc.arg('plano_id'),
    sqlc.arg('status'),
    sqlc.arg('mensalidade'),
    sqlc.arg('vencimento'),
    sqlc.arg('observacoes'),
    sqlc.arg('subdominio')
);