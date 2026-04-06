-- 1. Adiciona a nova coluna de ID (permitindo nulo temporariamente para não travar se houver dados)
ALTER TABLE entrada_nf ADD COLUMN Idfornecedor INT;

-- 2. Adiciona a Constraint de Chave Estrangeira
ALTER TABLE entrada_nf 
ADD CONSTRAINT fk_entrada_nf_fornecedor 
FOREIGN KEY (Idfornecedor) REFERENCES fornecedores(id);

-- 3. (OPCIONAL) Se você já tem nomes de fornecedores na coluna antiga, 
-- você precisaria fazer um UPDATE relacionando os nomes aos IDs da tabela de fornecedores aqui.

-- 4. Remove a "Regra de Ouro" antiga (Unique Constraint) que usava o nome em texto
ALTER TABLE entrada_nf DROP CONSTRAINT uk_entrada_nf_fornecedor_tenant;

-- 5. Remove a coluna de texto antiga
ALTER TABLE entrada_nf DROP COLUMN fornecedor;

-- 6. Torna a nova coluna obrigatória (NOT NULL) após a limpeza
ALTER TABLE entrada_nf ALTER COLUMN Idfornecedor SET NOT NULL;

-- 7. Cria a nova "Regra de Ouro" usando o ID do fornecedor
ALTER TABLE entrada_nf 
ADD CONSTRAINT uk_entrada_nf_fornecedor_tenant 
UNIQUE (tenant_id, Idfornecedor, nota_fiscal_numero, nota_fiscal_serie);