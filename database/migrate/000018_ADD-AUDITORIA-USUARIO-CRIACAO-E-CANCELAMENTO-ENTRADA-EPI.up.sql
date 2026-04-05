-- 1. Adicionando campos de auditoria na tabela de ITENS
ALTER TABLE entrada_epi_item 
ADD COLUMN id_usuario_criacao INT NOT NULL,
ADD COLUMN id_usuario_cancelamento INT NULL;

-- 2. Adicionando as Foreign Keys para garantir a integridade (opcional, mas recomendado)
ALTER TABLE entrada_epi_item 
ADD CONSTRAINT fk_usuario_criacao_item FOREIGN KEY (id_usuario_criacao) REFERENCES usuarios(id),
ADD CONSTRAINT fk_usuario_cancelamento_item FOREIGN KEY (id_usuario_cancelamento) REFERENCES usuarios(id);

-- 3. Adicionando campos de auditoria na tabela de NF (Cabeçalho)
-- É bom saber quem registrou a nota como um todo
ALTER TABLE entrada_nf 
ADD COLUMN id_usuario_criacao INT NOT NULL,
ADD COLUMN id_usuario_cancelamento INT NULL,
ADD COLUMN cancelada_em TIMESTAMP NULL;

ALTER TABLE entrada_nf 
ADD CONSTRAINT fk_usuario_criacao_nf FOREIGN KEY (id_usuario_criacao) REFERENCES usuarios(id),
ADD CONSTRAINT fk_usuario_cancelamento_nf FOREIGN KEY (id_usuario_cancelamento) REFERENCES usuarios(id);