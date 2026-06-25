-- name: GetTenantBySubdomain :one
SELECT id, nome_fantasia 
FROM empresas 
WHERE subdominio = $1 
  AND status IN ('Ativa', 'Em teste');

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
    sqlc.arg('vencimento'),
    sqlc.arg('observacoes'),
    sqlc.arg('subdominio')
);


-- name: EmpresasAtivas :one
select count(*) from empresas where status = 'Ativa';
-- name: EmpresaEmTeste :one
select count(*) from empresas where status = 'Em teste';
-- name: EmpresasBloqueadas :one
select count(*) from empresas where status = 'Bloqueada';   
-- name: TotalFuncionarios :one
SELECT COUNT(*) FROM funcionario;      
-- name: TotalEpis :one
SELECT COUNT(*) FROM epi; 
-- name: TotalEntregas :one
SELECT COUNT(*) FROM entrega_epi;
-- name: ReceitaMensal :one
SELECT COALESCE(SUM(mensalidade), 0)::float8 FROM planos;


-- name: EmpresasRecentes :many
SELECT 
    e.id, 
    e.nome_fantasia, 
    e.subdominio, 
    e.responsavel,
    e.status, 
    p.nome AS plano_nome, 
    p.mensalidade::float8,
    -- As subqueries agora estão no lugar certo (no SELECT), separadas por vírgula
    (SELECT COUNT(*)::int FROM funcionario f WHERE f.tenant_id = e.id) AS funcionarios,
    (SELECT COUNT(*)::int FROM epi ep WHERE ep.tenant_id = e.id) AS epis
FROM empresas e
INNER JOIN planos p ON e.plano_id = p.id;;


-- name: DadosEmpresas :many
SELECT 
    e.id, 
    e.nome_fantasia as nome, 
    e.cnpj,
    e.responsavel,
    e.email,
    e.telefone, 
    p.nome AS plano_nome, 
    (SELECT COUNT(*)::int FROM funcionario f WHERE f.tenant_id = e.id) AS funcionarios,
    (SELECT COUNT(*)::int FROM epi ep WHERE ep.tenant_id = e.id) AS epis,
    p.mensalidade::float8,
    e.vencimento,
    e.status    
FROM empresas e
INNER JOIN planos p ON e.plano_id = p.id;