ALTER TABLE epis_entregues DROP CONSTRAINT IF EXISTS fk_epis_entregues_cabecalho;

ALTER TABLE epis_entregues
ADD CONSTRAINT epis_entregues_id_entrega_cabecalho_fkey
FOREIGN KEY (id_entrega_cabecalho) REFERENCES entrada_nf(id);
