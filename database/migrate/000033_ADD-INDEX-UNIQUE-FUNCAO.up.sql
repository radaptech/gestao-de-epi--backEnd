ALTER TABLE funcao 
ADD CONSTRAINT uq_funcao_tenant_nome_departamento 
UNIQUE (tenant_id, nome, Iddepartamento);