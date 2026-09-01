-- Recria a coluna ativo a partir do status textual criado no up.
ALTER TABLE empresas ADD COLUMN ativo BOOLEAN NOT NULL DEFAULT TRUE;
UPDATE empresas SET ativo = (status = 'Ativa');

ALTER TABLE empresas
    DROP COLUMN IF EXISTS plano_id,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS mensalidade,
    DROP COLUMN IF EXISTS vencimento,
    DROP COLUMN IF EXISTS observacoes,
    DROP COLUMN IF EXISTS responsavel,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS telefone;
