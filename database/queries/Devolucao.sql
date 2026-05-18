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
select d.id, d.data_devolucao, d.idfuncionario, f.nome as funcionarioNome, f.matricula, e.nome as epiNome, 
t.tamanho, d.quantidadeadevolver, m.motivo, d.houve_troca, en.nome as epiNovo, tn.tamanho as tamanhoNovo,
d.quantidadenova,d.observacao, d.assinatura_digital, d.token_validacao
from devolucao d
inner join funcionario f on f.id = d.idfuncionario 
inner join epi e on e.id = d.idepi
full outer join epi en on en.id = d.idepinovo
inner join tamanhos_epis te on te.idtamanho = d.idtamanho
full outer join tamanho tn ON tn.id = d.idtamanhonovo
inner join tamanho t on t.id = te.idtamanho
inner join motivo_devolucao m on m.id = d.idmotivo
where d.tenant_id = $1 and d.cancelada_em is null
order by d.data_devolucao desc;

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