alter table funcionario
add column cpf VARCHAR(11),
add CONSTRAINT uk_Cpf_funcionario_tenant unique (tenant_id, cpf);