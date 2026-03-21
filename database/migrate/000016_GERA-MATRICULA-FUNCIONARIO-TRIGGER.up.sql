-- Cria a função que calcula o próximo número
CREATE OR REPLACE FUNCTION gerar_matricula_por_tenant()
RETURNS TRIGGER AS $$
BEGIN
    -- Olha para os funcionários do mesmo tenant, pega a maior matrícula e soma 1.
    -- Se for o primeiro funcionário (NULL), o COALESCE transforma em 0, e ele ganha a matrícula 1.
    SELECT COALESCE(MAX(matricula), 0) + 1
    INTO NEW.matricula
    FROM funcionario
    WHERE tenant_id = NEW.tenant_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Associa a função à tabela
CREATE TRIGGER trigger_gerar_matricula
BEFORE INSERT ON funcionario
FOR EACH ROW
-- Só executa se o Go não enviar uma matrícula manualmente
WHEN (NEW.matricula IS NULL) 
EXECUTE FUNCTION gerar_matricula_por_tenant();