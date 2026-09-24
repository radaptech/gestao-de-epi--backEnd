ALTER TABLE usuarios
  ALTER COLUMN role SET DEFAULT 'colaborador',
  ALTER COLUMN tenant_id SET NOT NULL;
