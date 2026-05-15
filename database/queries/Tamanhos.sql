-- name: AddTamanho :exec
INSERT INTO tamanho (tenant_id, tamanho) 
VALUES ($1, $2);

-- name: BuscarTamanho :one
SELECT id, tamanho 
FROM tamanho 
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE 
LIMIT 1;

-- name: BuscarTamanhosPorIdEpi :many
SELECT 
    t.id, 
    t.tamanho, 
    te.IdEpi,
    -- Soma o estoque disponível deste tamanho e EPI
    COALESCE((
        SELECT SUM(eei.quantidade_atual)
        FROM entrada_epi_item eei
        WHERE eei.id_epi = te.IdEpi 
          AND eei.id_tamanho = t.id
          AND eei.tenant_id = @tenant_id
          AND eei.ativo = TRUE
          AND eei.quantidade_atual > 0
          AND eei.data_validade >= CURRENT_DATE
    ), 0)::int AS saldo_atual
FROM tamanho t
INNER JOIN tamanhos_epis te ON t.id = te.IdTamanho
WHERE te.IdEpi = @id_epi 
  AND te.tenant_id = @tenant_id 
  AND te.ativo = TRUE
  AND t.tenant_id = @tenant_id 
  AND t.ativo = TRUE
ORDER BY t.tamanho ASC;


-- name: BuscarTamanhosComEstoquePorEpi :many
SELECT DISTINCT t.id, t.tamanho, eei.id_epi
FROM tamanho t
INNER JOIN entrada_epi_item eei ON t.id = eei.id_tamanho
WHERE eei.id_epi = $1 
  AND eei.tenant_id = $2 
  AND eei.quantidade_atual > 0
  AND eei.ativo = TRUE
ORDER BY t.tamanho ASC;
-- name: BuscarTodosTamanhos :many
SELECT id, tamanho 
FROM tamanho 
WHERE tenant_id = $1 -- SEGURANÇA: Lista apenas tamanhos desta empresa
  AND ativo = TRUE
ORDER BY tamanho ASC;

-- name: DeletarTamanho :execrows
UPDATE tamanho
SET ativo = FALSE,
    deletado_em = current_date
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE; -- SEGURANÇA: Garante que só pode deletar tamanhos da própria empresa

-- name: UpdateEpiNosTamanhos :execrows
-- Esta query atualiza a associação na tabela muitos-para-muitos
UPDATE tamanhos_epis
SET IdEpi = $2
WHERE IdEpi = $1 
  AND tenant_id = $3 -- SEGURANÇA: Obrigatório
  AND ativo = TRUE;

