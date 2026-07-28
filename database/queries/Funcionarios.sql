-- name: AddFuncionario :exec
INSERT INTO funcionario (tenant_id, nome, IdDepartamento, IdFuncao, cpf) 
VALUES ($1, $2, $3, $4, $5);

-- name: BuscaFuncionario :one
SELECT 
    fn.id, 
    fn.nome, 
    fn.matricula, 
    fn.cpf,
    fn.IdDepartamento, 
    d.nome as departamento_nome,
    fn.IdFuncao, 
    f.nome as funcao_nome
FROM funcionario fn
INNER JOIN departamento d ON fn.IdDepartamento = d.id
INNER JOIN funcao f ON fn.IdFuncao = f.id
WHERE fn.matricula = $1 
  AND fn.tenant_id = $2 -- IMPORTANTE: Matrícula só é única dentro do tenant
  AND fn.ativo = TRUE;

-- name: BuscarTodosFuncionarios :many
SELECT 
    fn.id, 
    fn.nome, 
    fn.cpf,
    -- Formatação para exibição
    CASE 
        WHEN fn.matricula < 10000 THEN LPAD(fn.matricula::text, 4, '0') 
        ELSE fn.matricula::text 
    END AS matricula,
    fn.IdDepartamento, 
    d.nome as departamento_nome,
    fn.IdFuncao, 
    f.nome as funcao_nome,
    COUNT(*) OVER() AS total_geral
FROM funcionario fn
INNER JOIN departamento d ON fn.IdDepartamento = d.id
INNER JOIN funcao f ON fn.IdFuncao = f.id
WHERE fn.tenant_id = sqlc.arg('tenant_id') 
  AND (sqlc.narg('id')::int IS NULL OR fn.id = sqlc.narg('id')::int)
  -- Formatação aplicada também na hora da busca para o ILIKE bater certinho
  AND (
      sqlc.narg('matricula')::text IS NULL OR 
      (
          CASE 
              WHEN fn.matricula < 10000 THEN LPAD(fn.matricula::text, 4, '0') 
              ELSE fn.matricula::text 
          END
      ) ILIKE '%' || sqlc.narg('matricula') || '%'
  )
  AND (sqlc.narg('nome')::text IS NULL OR fn.nome ILIKE '%' || sqlc.narg('nome') || '%')
  AND (
    (sqlc.arg('cancelados')::boolean IS FALSE AND fn.deletado_em IS NULL) OR
    (sqlc.arg('cancelados')::boolean IS TRUE AND fn.deletado_em IS NOT NULL)
  )
ORDER BY fn.nome ASC
LIMIT $1 OFFSET $2;

-- name: DeletarFuncionario :execrows
UPDATE funcionario
SET ativo = FALSE,
    deletado_em = current_date
WHERE id = $1 
  AND tenant_id = $2 -- SEGURANÇA
  AND ativo = TRUE;

-- name: UpdateFuncionarioNome :execrows
UPDATE funcionario
SET nome = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND ativo = TRUE;

-- name: UpdateFuncionarioCpf :execrows
UPDATE funcionario
SET cpf = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND ativo = TRUE;

-- name: UpdateFuncionarioDepartamento :execrows
UPDATE funcionario
SET IdDepartamento = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND ativo = TRUE;

-- name: UpdateFuncionarioFuncao :execrows
UPDATE funcionario
SET IdFuncao = $2
WHERE id = $1 
  AND tenant_id = $3 -- SEGURANÇA
  AND ativo = TRUE;


-- name: BuscaFuncionarioPorId :one
SELECT 
    fn.id, 
    fn.nome, 
    fn.matricula, 
    fn.IdDepartamento, 
    d.nome as departamento_nome,
    fn.IdFuncao, 
    f.nome as funcao_nome
FROM funcionario fn
INNER JOIN departamento d ON fn.IdDepartamento = d.id
INNER JOIN funcao f ON fn.IdFuncao = f.id
WHERE fn.id = $1 
  AND fn.tenant_id = $2 -- SEGURANÇA
  AND fn.ativo = TRUE;

-- name: BuscaFuncionarioDashbord :many
SELECT id, nome, matricula, cpf
FROM funcionario
WHERE tenant_id = $1 
  AND ativo = TRUE 
  order by nome asc;

-- name: BuscaFuncionarioCompleto :many
SELECT
    fn.id, 
    fn.nome, 
    fn.matricula, 
    fn.IdFuncao, 
    fn.cpf,
    f.nome as funcao_nome,
    fn.IdDepartamento, 
    d.nome as departamento_nome
FROM funcionario fn
INNER JOIN departamento d ON fn.IdDepartamento = d.id
INNER JOIN funcao f ON fn.IdFuncao = f.id
WHERE fn.tenant_id = $1 
  AND fn.ativo = TRUE;

-- name: TotalDeFuncionarios :one
select count(id) from funcionario where tenant_id= @id;