ALTER TABLE usuarios
DROP COLUMN IF EXISTS token_recuperacao_senha,
DROP COLUMN IF EXISTS token_expiracao;
