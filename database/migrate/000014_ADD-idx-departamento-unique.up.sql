CREATE UNIQUE INDEX idx_departamento_nome 
ON departamento (tenant_id, nome) 
WHERE deletado_em IS NULL;