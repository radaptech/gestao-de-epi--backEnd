# Referência: helpers e construtores para testes de integração

Todos em `internal/service/` (pacote `service`), arquivos `*_test.go`.
Confira o código-fonte antes de usar — esta lista pode ficar desatualizada
se os helpers mudarem.

## Setup base

```go
db := SetupTestDB(t)              // *pgxpool.Pool, container Postgres real + migrations reais
defer db.Close()
```

## Helpers de fixture (`SetupDadosTesteIntegracao_test.go`)

Todos retornam `int64` (o ID criado) e recebem `*testing.T` + `*pgxpool.Pool`
como primeiros parâmetros.

| Helper | Assinatura (sem t, db) | Observação |
|---|---|---|
| `CreatePlanos` | `()` | plano de teste com limites de 1000 |
| `CreateEmpresa` | `(idplanos int64)` | gera CNPJ/subdomínio únicos por timestamp |
| `CreateUser` | `(tenantID int64)` | role fixo `"admin"` |
| `CreateDepartamento` | `(tenantID int64)` | nome único via timestamp |
| `CreateFuncao` | `(idDep, tenantID int64)` | |
| `CreateFuncionario` | `(IdDepartamento, IdFuncao, tenantID int64)` | matrícula aleatória, CPF único (11 dígitos) |
| `CreateProtecao` | `(tenantID int64)` | tipo_protecao |
| `CreateTamanho` | `(tenantID int64)` | |
| `CreateEpi` | `(idTipoProtecao, tenantID int64)` | CA aleatório, validade +1 ano |
| `CreateFornecedor` | `(tenantID int64)` | CNPJ aleatório de 14 dígitos |
| `CreateEntradaNfEpi` | `(tenantID, iduser, idfornecedor int64)` | cabeçalho da NF |
| `CreateEntradaEpi` | `(tenantid, idEpi, idTamanho, iduser, IDentradaNf int64)` | lote com **100 unidades**, validade +3 anos |
| `CreateEntradaEpi1` | mesma assinatura | variante com **1 unidade** só (para testes de concorrência forçando esgotamento) |
| `CreateEntregaEpi` | `(idFuncionario, idUserEntrega, tenantID int64)` | INSERT cru do cabeçalho `entrega_epi` — não abate estoque |
| `CreateEpiEntregues` | `(IDEntregaCabecalho, idEntradaItem, IdEpi, IdTamanho, tenantID int64)` | INSERT cru de `epis_entregues`, quantidade fixa 10 — **não decrementa `entrada_epi_item.quantidade_atual`**; se precisar de saldo real, use `EntregaService.Salvar` de verdade em vez desses dois |
| `CreateMotivoDevolucao` | `(motivo string, geraDescarte bool, tenantID int64)` | |

## Construtores de repository (`database/repository`)

Todos recebem `*pgxpool.Pool` e devolvem o ponteiro do repo.

```
repository.NewDepartamentoRepository(pool)
repository.NewFuncaoRepository(pool)
repository.NewFuncionarioRepository(pool)
repository.NewTamanhoRepository(pool)
repository.NewProtecaoRepository(pool)        // tipo_protecao
repository.NewEpiRepository(pool)
repository.NewEntradaRepository(pool)
repository.NewFornecedorRepository(pool)
repository.NewEntregaRepository(pool)
repository.NewEstoqueRepository(pool)
repository.NewMotivoDevolucaoRepository(pool)
repository.NewDevolucaoRepository(pool)
repository.NewPlanosRepository(pool)
repository.NewEmpresaRepository(pool)
repository.NewUsuarioRepository(pool)
```

## Construtores de service (`internal/service`)

```
service.NewDepartamentoService(repo)
service.NewFuncaoService(repo)
service.NewFuncionarioService(repoFuncionario, repoEntrega, repoEpi, db)
service.NewTamanhoService(repo)
service.NewProtecaoService(repo)
service.NewEpiService(repo, db)
service.NewEntradaService(repo, db)
service.NewFornecedorService(repo)
service.NewEntregaService(repo, db)
service.NewEstoqueService(repo)
service.NewMotivoDevolucaoRepositoryServe(repo)   // nome não segue o padrão "Service" no final
service.NewDevolucaoService(repoDevolucao, db, *entregaService, repoMotivo) // entregaService POR VALOR, não ponteiro
service.NewPlanoService(repo)
service.NewEmpresaService(repoEmpresas, repoPlanos)
service.NewUsuarioService(repo, *emailService, db)
```

## Exemplo mínimo de esqueleto de teste

```go
package service

import (
	"context"
	"testing"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/stretchr/testify/require"
)

func TestEpi(t *testing.T) {
	t.Run("nome do cenario", func(t *testing.T) {
		ctx := context.Background()
		db := SetupTestDB(t)
		defer db.Close()

		planos := CreatePlanos(t, db)
		empresa := CreateEmpresa(t, db, planos)
		idprotec := CreateProtecao(t, db, empresa)

		repo := repository.NewEpiRepository(db)
		serv := NewEpiService(repo, db)

		// ... exercitar serv, e require.Equal/require.NoError no resultado
		_ = ctx
		_ = serv
	})
}
```
