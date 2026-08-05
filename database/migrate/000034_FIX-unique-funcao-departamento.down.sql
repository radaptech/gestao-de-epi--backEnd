DROP INDEX IF EXISTS idx_funcao_nome_tenant_departamento_ativo;

CREATE UNIQUE INDEX idx_funcao_nome_tenant_ativo
ON funcao (tenant_id, nome)
WHERE deletado_em IS NULL;

ALTER TABLE funcao
ADD CONSTRAINT uq_funcao_tenant_nome_departamento
UNIQUE (tenant_id, nome, Iddepartamento);
