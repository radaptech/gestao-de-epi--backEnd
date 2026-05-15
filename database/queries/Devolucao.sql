-- name: AddDevolucaoSimples :exec
INSERT INTO devolucao (
    tenant_id, IdFuncionario, IdEpi, IdMotivo, data_devolucao, IdTamanho, 
    quantidadeAdevolver, assinatura_digital, id_usuario_cancelamento, token_validacao
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: AddTrocaEpi :one
INSERT INTO devolucao (
    tenant_id, IdFuncionario, IdEpi, IdMotivo, data_devolucao, IdTamanho, 
    quantidadeAdevolver, IdEpiNovo, IdTamanhoNovo, quantidadeNova, assinatura_digital, token_validacao
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id;

-- name: AddEntregaVinculada :one
INSERT INTO entrega_epi (tenant_id, IdFuncionario, data_entrega, assinatura, IdTroca)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: ListarDevolucoes :many
SELECT 
    d.id, d.IdFuncionario, f.nome as func_nome, f.matricula,
    f.IdDepartamento, dd.nome as dep_nome,
    f.IdFuncao, ff.nome as funcao_nome,
    d.IdEpi, e.nome as epi_antigo_nome, e.fabricante as epi_antigo_fab, e.CA as epi_antigo_ca,
    d.IdTamanho as tam_antigo_id, t.tamanho as tam_antigo_nome, e.descricao as desc_antiga,
    e.validade_CA as validade_ca_antiga, e.IdTipoProtecao as idprotecaoAntigo, tp.nome as tipo_protecao_nomeAntigo,
    d.quantidadeAdevolver, d.IdMotivo, m.motivo as motivo_nome,
    d.IdEpiNovo, 
    en.nome as epi_novo_nome, en.fabricante as epi_novo_fab, en.CA as epi_novo_ca,
    d.quantidadeNova, d.IdTamanhoNovo, tn.tamanho as tam_novo_nome, en.descricao as desc_nova,
    en.validade_CA as validade_ca_nova, en.IdTipoProtecao as idprotecaoNovo, tpn.nome as tipo_protecao_nomeNovo,
    d.assinatura_digital, d.data_devolucao, d.id_usuario_cancelamento,
    COUNT(*) OVER() as total_geral
FROM devolucao d
INNER JOIN epi e ON d.IdEpi = e.id
INNER JOIN funcionario f ON d.IdFuncionario = f.id  
INNER JOIN departamento dd ON f.IdDepartamento = dd.id
INNER JOIN funcao ff ON f.IdFuncao = ff.id
INNER JOIN tamanho t ON d.IdTamanho = t.id
INNER JOIN motivo_devolucao m ON d.IdMotivo = m.id
INNER JOIN tipo_protecao tp ON e.IdTipoProtecao = tp.id
LEFT JOIN epi en ON d.IdEpiNovo = en.id
LEFT JOIN tamanho tn ON d.IdTamanhoNovo = tn.id
LEFT JOIN tipo_protecao tpn ON en.IdTipoProtecao = tpn.id
WHERE 
    d.tenant_id = sqlc.arg('tenant_id') -- SEGURANÇA: Filtro obrigatório do tenant
    AND ((sqlc.arg('canceladas')::boolean IS FALSE AND d.cancelada_em IS NULL) OR
         (sqlc.arg('canceladas')::boolean IS TRUE AND d.cancelada_em IS NOT NULL))
    AND (sqlc.narg('id')::int IS NULL OR d.id = sqlc.narg('id'))
    AND (sqlc.narg('matricula')::text IS NULL OR f.matricula = sqlc.narg('matricula'))
    AND (sqlc.narg('data_inicio')::date IS NULL OR d.data_devolucao >= sqlc.narg('data_inicio'))
    AND (sqlc.narg('data_fim')::date IS NULL OR d.data_devolucao <= sqlc.narg('data_fim'))
ORDER BY d.data_devolucao DESC 
LIMIT $1 OFFSET $2;

-- name: CancelarDevolucao :one
UPDATE devolucao
SET cancelada_em = current_date,
    ativo = FALSE,
    id_usuario_devolucao_cancelamento = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA: Só cancela se pertencer à empresa correta
  AND cancelada_em IS NULL
RETURNING id;


-- name: ConsultarSaldoEpiFuncionario :one
SELECT
    (
        -- Total já entregue para este funcionário
        COALESCE((
            SELECT SUM(eei.quantidade)::int
            FROM epis_entregues eei
            JOIN entrega_epi ee ON ee.id = eei.id_entrega_cabecalho
            WHERE ee.idfuncionario = $1
              AND eei.id_epi = $2
              AND eei.id_tamanho = $3
              AND ee.tenant_id = $4
        ), 0)
        -
        -- Menos o total que ele já devolveu anteriormente
        COALESCE((
            SELECT SUM(d.quantidadeadevolver)::int
            FROM devolucao d
            WHERE d.idfuncionario = $1
              AND d.idepi = $2
              AND d.idtamanho = $3
              AND d.tenant_id = $4
        ), 0)
    )::int AS saldo;


-- name: ListarLotesParaRepor :many
-- Busca todos os lotes que ainda têm espaço para receber devolução (quantidade_atual < quantidade)
SELECT id, quantidade, quantidade_atual 
FROM entrada_epi_item
WHERE tenant_id = $1 
  AND id_epi = $2 
  AND id_tamanho = $3
  AND ativo = TRUE
  AND quantidade_atual < quantidade
ORDER BY id DESC; -- Pega do lote mais recente para o mais antigo

-- name: AtualizarSaldoLote :exec
-- Atualiza o saldo de um lote específico
UPDATE entrada_epi_item
SET quantidade_atual = $2
WHERE id = $1 AND tenant_id = $3;