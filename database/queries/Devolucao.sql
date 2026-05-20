-- name: AddDevolucaoSimples :exec
INSERT INTO devolucao (
    tenant_id, IdFuncionario, IdEpi, IdMotivo, data_devolucao, IdTamanho, 
    quantidadeAdevolver, assinatura_digital, id_usuario_cancelamento, token_validacao
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: AddTrocaEpi :one
INSERT INTO devolucao (
    tenant_id, IdFuncionario, IdEpi, IdMotivo, data_devolucao, IdTamanho, 
    quantidadeAdevolver, IdEpiNovo, IdTamanhoNovo, quantidadeNova, assinatura_digital,id_usuario_cancelamento, token_validacao, houve_troca, observacao
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING id;

-- name: AddEntregaVinculada :one
INSERT INTO entrega_epi (tenant_id, IdFuncionario, data_entrega, assinatura, IdTroca)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: ListarDevolucoes :many
SELECT 
    d.id, 
    d.data_devolucao, 
    d.idfuncionario, 
    f.nome AS funcionarioNome, 
    f.matricula, 
    e.nome AS epiNome, 
    t.tamanho, 
    d.quantidadeadevolver, 
    m.motivo, 
    d.houve_troca, 
    en.nome AS epiNovo, 
    tn.tamanho AS tamanhoNovo,
    d.quantidadenova,
    d.observacao, 
    d.assinatura_digital, 
    d.token_validacao
FROM devolucao d
INNER JOIN funcionario f ON f.id = d.idfuncionario 
INNER JOIN epi e ON e.id = d.idepi
INNER JOIN tamanho t ON t.id = d.idtamanho            
INNER JOIN motivo_devolucao m ON m.id = d.idmotivo
LEFT JOIN epi en ON en.id = d.idepinovo               
LEFT JOIN tamanho tn ON tn.id = d.idtamanhonovo       
WHERE d.tenant_id = $1 AND d.cancelada_em IS NULL
ORDER BY d.data_devolucao DESC;


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

-- name: BuscarDadosPdfDevolucao :one
SELECT 
    em.razao_social AS nome_empresa, 
    f.nome AS funcionario_nome, 
    f.matricula, 
    dep.nome AS setor, 
    ff.nome AS cargo, 
    d.assinatura_digital,
    d.data_devolucao, 
    d.quantidadeadevolver, 
    e.nome AS epi_nome, 
    t.tamanho, 
    m.motivo, 
    d.houve_troca,               
    en.nome AS epi_novo, 
    tn.tamanho AS tamanho_novo, 
    d.quantidadenova
FROM devolucao d
INNER JOIN empresas em ON em.id = d.tenant_id
INNER JOIN funcionario f ON f.id = d.idfuncionario
INNER JOIN funcao ff ON ff.id = f.idfuncao
INNER JOIN departamento dep ON dep.id = f.iddepartamento
INNER JOIN epi e ON e.id = d.idepi
INNER JOIN tamanho t ON t.id = d.idtamanho
INNER JOIN motivo_devolucao m ON m.id = d.idmotivo
LEFT JOIN epi en ON en.id = d.idepinovo      
LEFT JOIN tamanho tn ON tn.id = d.idtamanhonovo 
WHERE d.id = $1 AND d.tenant_id = $2;