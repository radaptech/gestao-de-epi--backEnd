alter table usuarios
add column token_recuperacao_senha text,
add column token_expiracao timestamp;