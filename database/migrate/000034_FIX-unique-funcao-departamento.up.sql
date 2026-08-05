-- Corrige duas constraints de unicidade sobrepostas em `funcao`:
--   - idx_funcao_nome_tenant_ativo (000007): única por tenant inteiro, ignora departamento.
--   - uq_funcao_tenant_nome_departamento (000033): única por tenant+nome+departamento,
--     mas sem filtro de deletado_em, então nunca libera o nome mesmo após soft-delete.
-- Substitui as duas por um único índice parcial: nome único por departamento, liberado
-- para reuso quando o registro antigo está soft-deletado.

ALTER TABLE funcao DROP CONSTRAINT IF EXISTS uq_funcao_tenant_nome_departamento;
DROP INDEX IF EXISTS idx_funcao_nome_tenant_ativo;

CREATE UNIQUE INDEX idx_funcao_nome_tenant_departamento_ativo
ON funcao (tenant_id, nome, Iddepartamento)
WHERE deletado_em IS NULL;
