ALTER TABLE entrada_nf DROP CONSTRAINT IF EXISTS uk_entrada_nf_fornecedor_tenant;
ALTER TABLE entrada_nf DROP CONSTRAINT IF EXISTS fk_entrada_nf_fornecedor;

-- Volta a coluna de texto. O nome do fornecedor foi perdido no up, então as
-- linhas existentes ficam com string vazia (a coluna original era NOT NULL).
ALTER TABLE entrada_nf ADD COLUMN fornecedor VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE entrada_nf ALTER COLUMN fornecedor DROP DEFAULT;

ALTER TABLE entrada_nf DROP COLUMN IF EXISTS Idfornecedor;

ALTER TABLE entrada_nf
ADD CONSTRAINT uk_entrada_nf_fornecedor_tenant
UNIQUE (tenant_id, nota_fiscal_numero, nota_fiscal_serie, fornecedor);
