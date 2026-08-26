ALTER TABLE entrada_epi
DROP COLUMN IF EXISTS id_usuario_criacao,
DROP COLUMN IF EXISTS id_usuario_criacao_cancelamento;

ALTER TABLE entrega_epi
DROP COLUMN IF EXISTS id_usuario_entrega,
DROP COLUMN IF EXISTS id_usuario_entrega_cancelamento;

ALTER TABLE devolucao
DROP COLUMN IF EXISTS id_usuario_cancelamento,
DROP COLUMN IF EXISTS id_usuario_devolucao_cancelamento;
