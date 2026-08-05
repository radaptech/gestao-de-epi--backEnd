---
name: db-test-fixture
description: Gera testes de integração em internal/service/ seguindo o padrão real deste projeto (testcontainers com Postgres real, migrations reais, helpers Create*, um container isolado por subteste). Use ao criar testes de integração para um service que ainda não tem (ex. EpiService, FuncionarioService, FornecedorService, EmpresaService), ou ao revisar/corrigir um teste de integração existente.
tools: Read, Grep, Write, Edit, Bash
---

# DB Test Fixture

Este projeto testa a camada `internal/service/` com **Postgres real dentro
de Docker** via testcontainers-go — não com mocks. O padrão já existe nos
arquivos `internal/service/testeIntegracaoEntregaEpi_test.go` e
`testeIntegracaoDevolucaoEpi_test.go`; use-os como referência viva antes de
escrever um teste novo. Veja `references/fixtures.md` para a lista completa
de helpers e construtores já disponíveis.

## Peças do padrão

1. **`SetupTestDB(t) *pgxpool.Pool`** (`SetupTesteIntegracao_test.go`) — sobe
   um container Postgres 15-alpine via testcontainers, aplica as migrations
   REAIS de `database/migrate` (não um schema copiado à mão — isso já
   causou drift no passado) e devolve o pool pronto. Registra
   `t.Cleanup` para derrubar o container sozinho.
2. **Helpers `CreateX`** (`SetupDadosTesteIntegracao_test.go`) — inserem
   fixtures diretas (empresa, usuário, departamento, função, funcionário,
   proteção, tamanho, EPI, fornecedor, entrada de NF/lote). Ver referência.
3. Construa o `service` real com seu repository real
   (`repository.NewXRepository(db)` + `service.NewXService(repo, db)`),
   nunca mock — é assim que o resto da suíte funciona.

## Passo a passo para criar um teste novo

1. Nomeie o arquivo `testeIntegracao<Entidade>_test.go` (sufixo `_test.go`
   é obrigatório — sem ele o código vaza para o binário de produção, isso
   já aconteceu).
2. Uma função `TestX(t *testing.T)` no topo, com subtestes via `t.Run`.
3. **Cada subteste chama `SetupTestDB(t)` de novo e cria seu PRÓPRIO
   tenant/empresa do zero.** Não compartilhe entidades criadas fora dos
   `t.Run` entre subtestes — se dois subtestes usarem containers/pools
   diferentes mas uma variável de fora referenciar IDs de outro container,
   os testes só passam por coincidência de numeração de SERIAL, e quebram
   silenciosamente na primeira mudança de ordem de criação (já aconteceu,
   foi corrigido).
4. Se o teste precisa de um estado que só a lógica real do service produz
   corretamente (ex: estoque abatido), **use o service de verdade para
   criar esse estado**, não um INSERT cru. Ex: para simular "funcionário já
   recebeu N unidades", chame `EntregaService.Salvar(...)` de verdade — os
   helpers crus como `CreateEpiEntregues` não decrementam
   `entrada_epi_item.quantidade_atual`, então testes que dependem de saldo
   real do lote (ex: devolução repondo estoque) vão falhar de forma confusa
   se só usarem os helpers crus.
5. Sempre popule campos ponteiro opcionais dos DTOs (`*string`, `*int`) com
   um valor, mesmo em cenário de "sucesso" — a menos que o teste seja
   especificamente sobre o caminho nil (nesse caso, teste o nil de
   propósito: já achamos um `panic` de produção assim, em
   `DevolucaoService.SalvarDevolucao` com `Observacao` nil).
6. Use `require` (testify) para toda asserção, com mensagem explicando o
   valor esperado. Prefira números concretos e verificáveis (ex: lote
   começa com 100, então depois de devolver 4 deve estar em 94) a só
   `require.NoError` — um teste que só confere "não deu erro" não prova que
   o efeito colateral (estoque, cancelamento, etc.) aconteceu certo.
7. Feche o pool no fim do subteste com `defer base.db.Close()` (ou
   equivalente) por limpeza, mesmo que o `t.Cleanup` do container já
   garanta que ele morre.

## Armadilha de infraestrutura (só relevante se você mexer no MEIO de teste, não em testes novos)

Se `criarTabelasPostgres` (que roda `golang-migrate` contra o container)
precisar ser tocado: o driver do golang-migrate mantém um advisory lock do
Postgres numa conexão dedicada, emprestada do `pgxpool.Pool` via
`stdlib.OpenDBFromPool`. Fechar só o `*sql.DB` não libera essa conexão —
é preciso chamar `m.Close()` no `*migrate.Migrate`, senão `pool.Close()` no
fim do teste trava para sempre esperando essa conexão nunca devolvida.

## Rodando

```bash
go test -v ./internal/service/... -run TestNomeDaFuncao -p 1
```

Precisa do Docker rodando. `-p 1` é obrigatório (testcontainers não tolera
pacotes em paralelo). Costuma passar dos 2-3 minutos — rode em background
se estiver usando um wrapper que tem timeout curto.
