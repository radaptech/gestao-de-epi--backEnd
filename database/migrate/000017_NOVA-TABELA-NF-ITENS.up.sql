DROP table IF EXISTS epis_entregues CASCADE;
DROP TABLE entrada_epi CASCADE;


CREATE TABLE entrada_nf (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    fornecedor VARCHAR(100) NOT NULL,
    nota_fiscal_numero VARCHAR(50) NOT NULL,
    nota_fiscal_serie VARCHAR(10) DEFAULT '1',
    data_emissao DATE NOT NULL,
    data_registro TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    
    FOREIGN KEY (tenant_id) REFERENCES empresas(id),
    
    -- A regra de ouro: Mesma NF + Série + Fornecedor não entra duas vezes para o mesmo cliente
    CONSTRAINT uk_entrada_nf_fornecedor_tenant 
    UNIQUE (tenant_id, nota_fiscal_numero, nota_fiscal_serie, fornecedor)
);

CREATE TABLE entrada_epi_item (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    entrada_nf_id INT NOT NULL, -- Link com o cabeçalho acima
    id_epi INT NOT NULL,
    id_tamanho INT NOT NULL,
    quantidade INT NOT NULL,
    quantidade_atual INT NOT NULL, -- Importante para o controle de saldo do lote
    data_fabricacao DATE NOT NULL,
    data_validade DATE NOT NULL,
    lote VARCHAR(50) NOT NULL,
    valor_unitario DECIMAL(10,2) NOT NULL,
    cancelada_em TIMESTAMP NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,

    FOREIGN KEY (tenant_id) REFERENCES empresas(id),
    FOREIGN KEY (entrada_nf_id) REFERENCES entrada_nf(id) ON DELETE CASCADE,
    FOREIGN KEY (id_epi) REFERENCES epi(id),
    FOREIGN KEY (id_tamanho) REFERENCES tamanho(id)
);



CREATE TABLE epis_entregues (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    id_entrega_cabecalho INT NOT NULL, -- FK para a tabela que diz QUEM recebeu e QUANDO
    id_entrada_item INT NOT NULL,      -- FK para entrada_epi_item (O lote de onde saiu)
    id_epi INT NOT NULL,
    id_tamanho INT NOT NULL,
    quantidade INT NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    deletado_em TIMESTAMP NULL,

    FOREIGN KEY (tenant_id) REFERENCES empresas(id),
    FOREIGN KEY (id_entrega_cabecalho) REFERENCES entrada_nf(id),
    FOREIGN KEY (id_epi) REFERENCES epi(id),
    FOREIGN KEY (id_entrada_item) REFERENCES entrada_epi_item(id) -- AQUI A MUDANÇA
);