    -- 1. Remove a constraint que aponta para a tabela errada (entrada_nf)
ALTER TABLE epis_entregues DROP CONSTRAINT IF EXISTS epis_entregues_id_entrega_cabecalho_fkey;

-- 2. Cria a constraint apontando para a tabela correta (entrega_epi)
ALTER TABLE epis_entregues 
ADD CONSTRAINT fk_epis_entregues_cabecalho 
FOREIGN KEY (id_entrega_cabecalho) REFERENCES entrega_epi(id);