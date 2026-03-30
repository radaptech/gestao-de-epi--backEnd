-- name: AddEntradaEpi :exec
INSERT INTO entrada_epi (
    tenant_id, -- Novo campo obrigatório
    IdEpi, IdTamanho, data_entrada, quantidade, quantidadeAtual, 
    data_fabricacao, data_validade, lote, Idfornecedor, valor_unitario, nota_fiscal_numero, nota_fiscal_serie, id_usuario_criacao
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: ListarEntradas :many
SELECT 
    ee.id, 
    ee.IdEpi, 
    e.nome as epi_nome, 
    e.fabricante, 
    e.CA, 
    e.descricao as epi_descricao,
    ee.data_fabricacao, 
    ee.data_validade, 
    e.validade_CA,
    e.IdTipoProtecao, 
    tp.nome as protecao_nome,
    ee.IdTamanho, 
    t.tamanho as tamanho_nome, 
    ee.quantidade, 
    ee.quantidadeAtual, 
    ee.data_entrada,
    ee.lote, 
    ee.Idfornecedor,
    f.razao_social,
    f.nome_fantasia,
    f.cnpj,
    f.inscricao_estadual,
    ee.valor_unitario, 
    ee.nota_fiscal_numero, 
    ee.nota_fiscal_serie, 
    
    -- Campos de Usuário Criação
    ee.id_usuario_criacao,
    u_criacao.nome as usuario_criacao_nome,
    
    -- Campos de Usuário Cancelamento
    ee.id_usuario_criacao_cancelamento,
    u_cancelamento.nome as usuario_cancelamento_nome,
    ee.cancelada_em -- É bom retornar a data também para saber quando foi

FROM entrada_epi ee
INNER JOIN epi e ON ee.IdEpi = e.id
INNER JOIN tipo_protecao tp ON e.IdTipoProtecao = tp.id
INNER JOIN tamanho t ON ee.IdTamanho = t.id
INNER JOIN fornecedores f on ee.IdFornecedor = f.id

-- JOIN 1: Quem criou a entrada
LEFT JOIN usuarios u_criacao ON ee.id_usuario_criacao = u_criacao.id

-- JOIN 2: Quem cancelou a entrada (só vai retornar dados se tiver sido cancelada)
LEFT JOIN usuarios u_cancelamento ON ee.id_usuario_criacao_cancelamento = u_cancelamento.id

WHERE 
    ee.tenant_id = sqlc.arg('tenant_id')
    AND (
        (sqlc.arg('canceladas')::boolean IS FALSE AND ee.cancelada_em IS NULL) OR
        (sqlc.arg('canceladas')::boolean IS TRUE AND ee.cancelada_em IS NOT NULL)
    )
    AND (sqlc.narg('id_epi')::int IS NULL OR ee.IdEpi = sqlc.narg('id_epi'))
    AND (sqlc.narg('id_entrada')::int IS NULL OR ee.id = sqlc.narg('id_entrada'))
    AND (sqlc.narg('data_inicio')::date IS NULL OR ee.data_entrada >= sqlc.narg('data_inicio'))
    AND (sqlc.narg('data_fim')::date IS NULL OR ee.data_entrada <= sqlc.narg('data_fim'))
    AND (sqlc.narg('nota_fiscal')::text IS NULL OR ee.nota_fiscal_numero ILIKE '%' || sqlc.narg('nota_fiscal') || '%')
ORDER BY ee.data_entrada DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CancelarEntrada :execrows
UPDATE entrada_epi 
SET 
    cancelada_em = NOW(), 
    ativo = FALSE,
    id_usuario_criacao_cancelamento = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA: Só cancela se for do mesmo tenant
  AND cancelada_em IS NULL 
  AND quantidadeAtual = quantidade;

-- name: ContarEntradasFiltradas :one
SELECT COUNT(*) 
FROM entrada_epi ee
WHERE 
    ee.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Filtro de Tenant
    AND (
        (sqlc.arg('canceladas')::boolean IS FALSE AND ee.cancelada_em IS NULL) OR
        (sqlc.arg('canceladas')::boolean IS TRUE AND ee.cancelada_em IS NOT NULL)
    )
    AND (sqlc.narg('id_epi')::int IS NULL OR ee.IdEpi = sqlc.narg('id_epi'))
    AND (sqlc.narg('id_entrada')::int IS NULL OR ee.id = sqlc.narg('id_entrada'))
    AND (sqlc.narg('data_inicio')::date IS NULL OR ee.data_entrada >= sqlc.narg('data_inicio'))
    AND (sqlc.narg('data_fim')::date IS NULL OR ee.data_entrada <= sqlc.narg('data_fim'))
    AND (sqlc.narg('nota_fiscal')::text IS NULL OR ee.nota_fiscal_numero ILIKE '%' || sqlc.narg('nota_fiscal') || '%');


-- name: EntradaDashbord :many
SELECT
    id,IdEpi, IdTamanho, quantidadeAtual, valor_unitario, quantidade,
    data_entrada, lote
FROM entrada_epi
WHERE tenant_id = $1
ORDER BY data_entrada DESC;


-- name: EntradaEpiEstoque :many
SELECT
    ee.id, ee.lote, ee.quantidade as quantidade_inicial, ee.quantidadeAtual as quantidade_atual,
    ee.valor_unitario, ee.data_validade,
    ee.IdTamanho, t.tamanho as tamanho_nome,
    ee.IdEpi, e.nome as epi_nome, e.fabricante, e.CA, e.descricao, e.validade_CA,e.alerta_minimo,
    e.IdTipoProtecao, tp.nome as protecao_nome
FROM entrada_epi ee
inner JOIN tamanho t on ee.IdTamanho = t.id
inner JOIN epi e on ee.IdEpi = e.id
inner join tipo_protecao tp on e.IdTipoProtecao = tp.id
WHERE ee.tenant_id = $1;