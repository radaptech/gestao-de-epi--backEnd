ALTER TABLE entrada_epi DROP CONSTRAINT IF EXISTS unique_entrada_Nf;

-- Restaura a constraint antiga (sem tenant_id) que a 000010 removeu.
ALTER TABLE entrada_epi
ADD CONSTRAINT uk_entrada_nf_fornecedor
UNIQUE (nota_fiscal_numero, nota_fiscal_serie, fornecedor);
