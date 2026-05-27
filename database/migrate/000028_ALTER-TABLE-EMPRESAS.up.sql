-- 1. Cria todas as colunas novas (deixamos sem NOT NULL por enquanto para não dar erro nos dados existentes)
ALTER TABLE empresas 
    ADD COLUMN plano_id INTEGER REFERENCES planos(id) ON DELETE SET NULL,
    ADD COLUMN status VARCHAR(20),
    ADD COLUMN mensalidade DECIMAL(10, 2),
    ADD COLUMN vencimento DATE,
    ADD COLUMN observacoes TEXT,
    ADD COLUMN responsavel VARCHAR(100),
    ADD COLUMN email VARCHAR(100),
    ADD COLUMN telefone VARCHAR(20);

-- 2. Migração de Dados: Transforma o ativo/inativo antigo no texto novo do React
UPDATE empresas SET status = 'Ativa' WHERE ativo = true;
UPDATE empresas SET status = 'Bloqueada' WHERE ativo = false;

-- 3. Agora que todos os dados antigos têm um status válido, aplicamos as regras de segurança
ALTER TABLE empresas ALTER COLUMN status SET NOT NULL;
ALTER TABLE empresas ALTER COLUMN status SET DEFAULT 'Em teste';

-- 4. Finalmente, com os dados a salvos, podemos apagar a coluna antiga
ALTER TABLE empresas DROP COLUMN ativo;