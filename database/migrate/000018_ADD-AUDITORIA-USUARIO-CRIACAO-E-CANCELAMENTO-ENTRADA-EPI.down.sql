ALTER TABLE entrada_nf
DROP CONSTRAINT IF EXISTS fk_usuario_criacao_nf,
DROP CONSTRAINT IF EXISTS fk_usuario_cancelamento_nf;

ALTER TABLE entrada_nf
DROP COLUMN IF EXISTS id_usuario_criacao,
DROP COLUMN IF EXISTS id_usuario_cancelamento,
DROP COLUMN IF EXISTS cancelada_em;

ALTER TABLE entrada_epi_item
DROP CONSTRAINT IF EXISTS fk_usuario_criacao_item,
DROP CONSTRAINT IF EXISTS fk_usuario_cancelamento_item;

ALTER TABLE entrada_epi_item
DROP COLUMN IF EXISTS id_usuario_criacao,
DROP COLUMN IF EXISTS id_usuario_cancelamento;
