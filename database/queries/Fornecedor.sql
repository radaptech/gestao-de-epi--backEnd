-- name: CriarFornecedor :exec
INSERT INTO fornecedores (
    tenant_id, 
    razao_social, 
    nome_fantasia, 
    cnpj, 
    inscricao_estadual 
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetFornecedor :one
SELECT * FROM fornecedores 
WHERE id = $1 AND tenant_id = $2 AND cancelado_em IS NULL;

-- name: ListarFornecedores :many
SELECT 
    id, 
    razao_social, 
    nome_fantasia, 
    cnpj, 
    inscricao_estadual, 
    ativo,
    count(*) OVER() AS total_items -- Isso retorna o total para paginação sem precisar de duas queries
FROM fornecedores
WHERE 
    tenant_id = sqlc.arg('tenant_id')
        AND (
        sqlc.narg('ignorar_filtro_cancelado')::boolean IS TRUE -- Nova condição para ignorar o filtro de cancelados
        OR
        (sqlc.arg('canceladas')::boolean IS FALSE AND cancelado_em IS NULL) OR
        (sqlc.arg('canceladas')::boolean IS TRUE AND cancelado_em IS NOT NULL)
    )
    
    AND (sqlc.narg('nome')::text IS NULL OR nome_fantasia ILIKE '%' || sqlc.narg('nome') || '%' OR razao_social ILIKE '%' || sqlc.narg('nome') || '%')
    AND (sqlc.narg('cnpj')::text IS NULL OR cnpj ILIKE '%' || sqlc.narg('cnpj') || '%')
ORDER BY id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AtualizarFornecedor :execrows
UPDATE fornecedores
SET 
    razao_social       = COALESCE(sqlc.narg('razao_social'), razao_social),
    nome_fantasia      = COALESCE(sqlc.narg('nome_fantasia'), nome_fantasia),
    cnpj               = COALESCE(sqlc.narg('cnpj'), cnpj),
    inscricao_estadual = COALESCE(sqlc.narg('inscricao_estadual'), inscricao_estadual)
WHERE 
    id = sqlc.arg('id') 
    AND tenant_id = sqlc.arg('tenant_id') 
    AND cancelado_em IS NULL;

-- name: DeletarFornecedor :execrows
UPDATE fornecedores
SET 
    ativo = FALSE,
    cancelado_em = current_date
WHERE id = $1 AND tenant_id = $2 AND cancelado_em IS NULL;