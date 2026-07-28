CREATE TABLE planos (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    mensalidade  DECIMAL(10,2) NOT NULL,
    limite_funcionarios INT, -- Se for NULL, é ilimitado
    limite_usuarios INT,     -- Se for NULL, é ilimitado
    limite_epis INT,         -- Se for NULL, é ilimitado
    status VARCHAR(20) DEFAULT 'Ativo', 
    descricao TEXT NOT NULL,
    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);