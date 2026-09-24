ALTER TABLE entrada_epi DROP CONSTRAINT IF EXISTS uk_entrada_nf_fornecedor;

ALTER TABLE entrada_epi
DROP COLUMN IF EXISTS nota_fiscal_numero,
DROP COLUMN IF EXISTS nota_fiscal_serie;
