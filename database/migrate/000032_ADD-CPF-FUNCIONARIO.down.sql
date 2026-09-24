ALTER TABLE funcionario DROP CONSTRAINT IF EXISTS uk_Cpf_funcionario_tenant;
ALTER TABLE funcionario DROP COLUMN IF EXISTS cpf;
