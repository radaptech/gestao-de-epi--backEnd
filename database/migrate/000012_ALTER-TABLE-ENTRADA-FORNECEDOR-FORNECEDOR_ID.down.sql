ALTER TABLE entrada_epi DROP CONSTRAINT IF EXISTS fk_entrada_fornecedor;

ALTER TABLE entrada_epi RENAME COLUMN Idfornecedor TO fornecedor;

-- Volta para texto. O nome original do fornecedor não existe mais no banco,
-- então as linhas existentes ficam com string vazia (coluna é NOT NULL).
ALTER TABLE entrada_epi
ALTER COLUMN fornecedor TYPE VARCHAR(100) USING COALESCE(fornecedor::VARCHAR, '');
