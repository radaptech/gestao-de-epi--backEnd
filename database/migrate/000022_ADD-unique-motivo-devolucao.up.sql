ALTER TABLE motivo_devolucao 
ADD CONSTRAINT uniqueMotivoContraints UNIQUE (motivo, tenant_id);