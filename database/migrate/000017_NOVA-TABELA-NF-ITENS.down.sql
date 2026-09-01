-- ATENÇÃO: migração destrutiva nos dois sentidos. A 000017 dropou entrada_epi e
-- epis_entregues com CASCADE; este down recria as tabelas VAZIAS no formato que
-- existia ao fim da 000016 (000001 + 000002 + 000005 + 000010 + 000012).

DROP TABLE IF EXISTS epis_entregues CASCADE;
DROP TABLE IF EXISTS entrada_epi_item CASCADE;
DROP TABLE IF EXISTS entrada_nf CASCADE;

CREATE TABLE entrada_epi (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    IdEpi INT NOT NULL,
    IdTamanho INT NOT NULL,
    data_entrada DATE NOT NULL,
    quantidade INT NOT NULL,
    quantidadeAtual INT NOT NULL,
    data_fabricacao DATE NOT NULL,
    data_validade DATE NOT NULL,
    lote VARCHAR(50) NOT NULL,
    Idfornecedor INTEGER NOT NULL,
    valor_unitario DECIMAL(10,2) NOT NULL,
    cancelada_em TIMESTAMP NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    nota_fiscal_numero VARCHAR(50) NOT NULL,
    nota_fiscal_serie VARCHAR(10) DEFAULT '1',
    id_usuario_criacao INTEGER REFERENCES usuarios(id),
    id_usuario_criacao_cancelamento INTEGER REFERENCES usuarios(id),
    FOREIGN KEY (tenant_id) REFERENCES empresas(id),
    FOREIGN KEY (IdEpi) REFERENCES epi(id),
    FOREIGN KEY (IdTamanho) REFERENCES tamanho(id),
    CONSTRAINT fk_entrada_fornecedor FOREIGN KEY (Idfornecedor) REFERENCES fornecedores(id),
    CONSTRAINT unique_entrada_Nf UNIQUE (tenant_id, Idfornecedor, nota_fiscal_numero, nota_fiscal_serie)
);

CREATE TABLE epis_entregues (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL,
    IdEntrega INT NOT NULL,
    IdEntrada INT NOT NULL,
    IdEpi INT NOT NULL,
    IdTamanho INT NOT NULL,
    quantidade INT NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    deletado_em TIMESTAMP NULL,
    FOREIGN KEY (tenant_id) REFERENCES empresas(id),
    FOREIGN KEY (IdEntrega) REFERENCES entrega_epi(id),
    FOREIGN KEY (IdEpi) REFERENCES epi(id),
    FOREIGN KEY (IdEntrada) REFERENCES entrada_epi(id)
);
