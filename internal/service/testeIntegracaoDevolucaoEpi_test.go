package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// baseDevolucao agrupa tudo que os testes de devolução precisam: um tenant
// completo, um lote de 100 unidades de um EPI e uma entrega REAL (via
// EntregaService, para abater o estoque de verdade) já feita para um
// funcionário — simulando "o funcionário já está com o EPI em mãos".
//
// Cada subteste chama setupBaseDevolucao com seu PRÓPRIO container/tenant
// (mesma lição do teste de entrega: nunca compartilhar entidades entre
// subtestes que rodam em containers/tenants diferentes).
type baseDevolucao struct {
	db            *pgxpool.Pool
	empresa       int64
	iduser        int64
	idfuncionario int64
	idtam         int64
	idepi         int64
	identradaEpi  int64 // lote do EPI original (começa com 100 un)
	devolucaoServ *DevolucaoService
	entregaServ   *EntregaService
}

func setupBaseDevolucao(t *testing.T, quantidadeEntregue int32) baseDevolucao {
	db := SetupTestDB(t)

	planos := CreatePlanos(t, db)
	empresa := CreateEmpresa(t, db, planos)
	iduser := CreateUser(t, db, empresa)
	iddep := CreateDepartamento(t, db, empresa)
	idfuncao := CreateFuncao(t, db, iddep, empresa)
	idfuncionario := CreateFuncionario(t, db, iddep, idfuncao, empresa)
	idprotec := CreateProtecao(t, db, empresa)
	idtam := CreateTamanho(t, db, empresa)
	idepi := CreateEpi(t, db, idprotec, empresa)
	idfornecedor := CreateFornecedor(t, db, empresa)
	identradaNf := CreateEntradaNfEpi(t, db, empresa, iduser, idfornecedor)
	identradaEpi := CreateEntradaEpi(t, db, empresa, idepi, idtam, iduser, identradaNf) // lote com 100 un

	repoEntrega := repository.NewEntregaRepository(db)
	entregaServ := NewEntregaService(repoEntrega, db)

	repoDevolucao := repository.NewDevolucaoRepository(db)
	repoMotivo := repository.NewMotivoDevolucaoRepository(db)
	devolucaoServ := NewDevolucaoService(repoDevolucao, db, *entregaServ, repoMotivo)

	// Entrega REAL (passa pelo EntregaService de verdade) para que o lote
	// fique com quantidade_atual < quantidade e sobre "espaço" para a devolução repor.
	err := entregaServ.Salvar(context.Background(), model.EntregaParaInserir{
		ID_funcionario:     int32(idfuncionario),
		Id_user:            int32(iduser),
		Data_entrega:       *configs.NewDataBrPtr(time.Now()),
		Assinatura_Digital: "assinatura-base",
		Itens: []model.ItemParaInserir{
			{ID_epi: int32(idepi), ID_tamanho: int32(idtam), Quantidade: quantidadeEntregue},
		},
	}, int32(empresa), "token-entrega-base")
	require.NoError(t, err, "setup: entrega base deveria ter funcionado")

	return baseDevolucao{
		db:            db,
		empresa:       empresa,
		iduser:        iduser,
		idfuncionario: idfuncionario,
		idtam:         idtam,
		idepi:         idepi,
		identradaEpi:  identradaEpi,
		devolucaoServ: devolucaoServ,
		entregaServ:   entregaServ,
	}
}

// estoqueAtual lê direto do banco o quantidade_atual de um lote.
func estoqueAtual(t *testing.T, db *pgxpool.Pool, idEntradaItem int64) int32 {
	var q int32
	err := db.QueryRow(context.Background(),
		"SELECT quantidade_atual FROM entrada_epi_item WHERE id = $1", idEntradaItem).Scan(&q)
	require.NoError(t, err)
	return q
}

func TestDevolucao(t *testing.T) {

	t.Run("sucesso ao devolver epi sem troca (motivo NAO gera descarte) - repoe estoque no lote", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 10) // entrega 10 unidades -> lote 100 -> 90
		defer base.db.Close()

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo sem descarte", false, base.empresa)

		require.Equal(t, int32(90), estoqueAtual(t, base.db, base.identradaEpi), "lote deveria estar com 90 apos a entrega base")

		obs := "devolucao de teste"
		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 4,
			Iduser:              int(base.iduser),
			AssinaturaDigital:   "assinatura-devolucao",
			Observacao:          &obs,
		}, int32(base.empresa), "token-devolucao-1")
		require.NoError(t, err, "devolucao dentro do saldo e sem descarte deveria funcionar")

		require.Equal(t, int32(94), estoqueAtual(t, base.db, base.identradaEpi), "as 4 unidades devolvidas deveriam voltar para o lote")
	})

	t.Run("sucesso ao devolver epi com motivo que gera descarte - NAO repoe estoque", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 10) // lote 100 -> 90
		defer base.db.Close()

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo com descarte", true, base.empresa)

		obs := "epi danificado, descartado"
		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 5,
			Iduser:              int(base.iduser),
			AssinaturaDigital:   "assinatura-devolucao",
			Observacao:          &obs,
		}, int32(base.empresa), "token-devolucao-2")
		require.NoError(t, err, "devolucao por descarte deveria funcionar mesmo sem repor estoque")

		require.Equal(t, int32(90), estoqueAtual(t, base.db, base.identradaEpi), "descarte nao deveria devolver nada ao lote")
	})

	t.Run("ERRO - tentar devolver mais unidades do que o funcionario tem em maos", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 5) // funcionario so tem 5 em maos
		defer base.db.Close()

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo qualquer", false, base.empresa)

		obs := "tentativa de devolucao acima do saldo"
		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 10, // so tem 5
			Iduser:              int(base.iduser),
			AssinaturaDigital:   "assinatura-devolucao",
			Observacao:          &obs,
		}, int32(base.empresa), "token-devolucao-erro-saldo")
		require.Error(t, err, "nao deveria permitir devolver mais do que o funcionario tem em maos")
		require.Contains(t, err.Error(), "possui apenas")

		var count int
		err = base.db.QueryRow(ctx, "SELECT count(*) FROM devolucao WHERE token_validacao = $1 AND tenant_id = $2",
			"token-devolucao-erro-saldo", base.empresa).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count, "a devolucao nao deveria ter sido persistida")

		require.Equal(t, int32(95), estoqueAtual(t, base.db, base.identradaEpi), "estoque nao deveria ter sido alterado")
	})

	t.Run("sucesso ao devolver com troca - gera nova entrega automaticamente e abate o estoque do epi novo", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 10) // lote do epi velho: 100 -> 90
		defer base.db.Close()

		// EPI novo, com seu proprio lote de 100 unidades, para receber a troca
		idprotecNovo := CreateProtecao(t, base.db, base.empresa)
		idEpiNovo := CreateEpi(t, base.db, idprotecNovo, base.empresa)
		idfornecedorNovo := CreateFornecedor(t, base.db, base.empresa)
		identradaNfNovo := CreateEntradaNfEpi(t, base.db, base.empresa, base.iduser, idfornecedorNovo)
		identradaEpiNovo := CreateEntradaEpi(t, base.db, base.empresa, idEpiNovo, base.idtam, base.iduser, identradaNfNovo) // 100 un

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo troca", false, base.empresa)

		idEpiNovoInt := int(idEpiNovo)
		idTamNovoInt := int(base.idtam)
		novaQtd := 3

		obs := "troca por tamanho errado"
		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 10, // devolve tudo que tinha
			Iduser:              int(base.iduser),
			Troca:               true,
			IdEpiNovo:           &idEpiNovoInt,
			IdTamanhoNovo:       &idTamNovoInt,
			NovaQuantidade:      &novaQtd,
			AssinaturaDigital:   "assinatura-troca",
			Observacao:          &obs,
		}, int32(base.empresa), "token-devolucao-troca")
		require.NoError(t, err, "devolucao com troca deveria funcionar")

		require.Equal(t, int32(100), estoqueAtual(t, base.db, base.identradaEpi), "o epi antigo deveria ter voltado 100% para o lote")
		require.Equal(t, int32(97), estoqueAtual(t, base.db, identradaEpiNovo), "a nova entrega gerada pela troca deveria ter abatido 3 unidades do epi novo")
	})

	t.Run("ERRO - troca sem informar o novo epi", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 10)
		defer base.db.Close()

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo troca invalida", false, base.empresa)

		obs := "troca sem novo epi"
		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 5,
			Iduser:              int(base.iduser),
			Troca:               true,
			IdEpiNovo:           nil, // faltando de proposito
			AssinaturaDigital:   "assinatura-devolucao",
			Observacao:          &obs,
		}, int32(base.empresa), "token-devolucao-troca-invalida")
		require.Error(t, err)
		require.Contains(t, err.Error(), "obrigatório para trocas")
	})

	t.Run("sucesso ao cancelar devolucao com troca - reverte a entrega vinculada gerada pela troca", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 10) // lote epi velho: 100 -> 90
		defer base.db.Close()

		idprotecNovo := CreateProtecao(t, base.db, base.empresa)
		idEpiNovo := CreateEpi(t, base.db, idprotecNovo, base.empresa)
		idfornecedorNovo := CreateFornecedor(t, base.db, base.empresa)
		identradaNfNovo := CreateEntradaNfEpi(t, base.db, base.empresa, base.iduser, idfornecedorNovo)
		identradaEpiNovo := CreateEntradaEpi(t, base.db, base.empresa, idEpiNovo, base.idtam, base.iduser, identradaNfNovo) // 100 un

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo troca cancelamento", false, base.empresa)

		idEpiNovoInt := int(idEpiNovo)
		idTamNovoInt := int(base.idtam)
		novaQtd := 3
		obs := "troca para depois cancelar"

		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 10,
			Iduser:              int(base.iduser),
			Troca:               true,
			IdEpiNovo:           &idEpiNovoInt,
			IdTamanhoNovo:       &idTamNovoInt,
			NovaQuantidade:      &novaQtd,
			AssinaturaDigital:   "assinatura-troca-cancelamento",
			Observacao:          &obs,
		}, int32(base.empresa), "token-devolucao-troca-cancelamento")
		require.NoError(t, err)

		require.Equal(t, int32(97), estoqueAtual(t, base.db, identradaEpiNovo), "a nova entrega da troca deveria ter abatido o lote novo antes do cancelamento")

		var idDevolucao int
		err = base.db.QueryRow(ctx, "SELECT id FROM devolucao WHERE token_validacao = $1 AND tenant_id = $2",
			"token-devolucao-troca-cancelamento", base.empresa).Scan(&idDevolucao)
		require.NoError(t, err)

		err = base.devolucaoServ.CancelarDevolucao(ctx, idDevolucao, int(base.iduser), int(base.empresa))
		require.NoError(t, err, "cancelamento da devolucao com troca deveria funcionar")

		require.Equal(t, int32(100), estoqueAtual(t, base.db, identradaEpiNovo),
			"cancelar a devolucao deveria reverter a entrega automatica da troca, devolvendo o lote novo ao estado original")

		// GAP CONHECIDO: CancelarDevolucao só reverte a entrega vinculada gerada pela
		// troca (via IdTroca). Ele NÃO reverte a reposição de estoque que a própria
		// devolução fez no lote do EPI antigo. Se este teste falhar porque o valor
		// abaixo virou 90, é porque esse comportamento foi corrigido — atualize o
		// valor esperado. Ver DevolucaoService.CancelarDevolucao.
		estoqueEpiAntigo := estoqueAtual(t, base.db, base.identradaEpi)
		fmt.Printf("⚠️  [GAP CONHECIDO] Estoque do EPI original apos cancelar a devolucao: %d (nao volta para 90)\n", estoqueEpiAntigo)
		require.Equal(t, int32(100), estoqueEpiAntigo, "documenta o comportamento atual: a reposicao da devolucao original NAO e revertida no cancelamento")
	})

	t.Run("sucesso ao devolver sem informar observacao (campo opcional, nil)", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 10) // lote 100 -> 90
		defer base.db.Close()

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo sem observacao", false, base.empresa)

		// Regressão: Observacao é *string opcional (sem binding:"required" no model);
		// o ShouldBindJSON deixa nil quando o campo é omitido do JSON. Já causou
		// panic (nil pointer dereference) em SalvarDevolucao antes de ser corrigido.
		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 3,
			Iduser:              int(base.iduser),
			AssinaturaDigital:   "assinatura-sem-observacao",
			Observacao:          nil,
		}, int32(base.empresa), "token-devolucao-sem-observacao")
		require.NoError(t, err, "devolucao sem observacao (nil) nao deveria dar panic nem erro")

		require.Equal(t, int32(93), estoqueAtual(t, base.db, base.identradaEpi))

		var observacao *string
		err = base.db.QueryRow(ctx, "SELECT observacao FROM devolucao WHERE token_validacao = $1 AND tenant_id = $2",
			"token-devolucao-sem-observacao", base.empresa).Scan(&observacao)
		require.NoError(t, err)
		require.Nil(t, observacao, "observacao nao informada deveria ser gravada como NULL")
	})

	t.Run("sucesso ao devolver com observacao preenchida - texto e persistido (nao vira NULL)", func(t *testing.T) {
		ctx := context.Background()
		base := setupBaseDevolucao(t, 10) // lote 100 -> 90
		defer base.db.Close()

		idMotivo := CreateMotivoDevolucao(t, base.db, "motivo com observacao", false, base.empresa)

		obs := "cabo do capacete rompeu"
		err := base.devolucaoServ.SalvarDevolucao(ctx, model.DevolucaoInserir{
			IdFuncionario:       int(base.idfuncionario),
			IdEpi:               int(base.idepi),
			IdMotivo:            int(idMotivo),
			IdTamanho:           int(base.idtam),
			DataDevolucao:       *configs.NewDataBrPtr(time.Now()),
			QuantidadeADevolver: 2,
			Iduser:              int(base.iduser),
			AssinaturaDigital:   "assinatura-com-observacao",
			Observacao:          &obs,
		}, int32(base.empresa), "token-devolucao-com-observacao")
		require.NoError(t, err)

		var observacaoSalva *string
		err = base.db.QueryRow(ctx, "SELECT observacao FROM devolucao WHERE token_validacao = $1 AND tenant_id = $2",
			"token-devolucao-com-observacao", base.empresa).Scan(&observacaoSalva)
		require.NoError(t, err)
		require.NotNil(t, observacaoSalva, "observacao preenchida nao deveria virar NULL")
		require.Equal(t, obs, *observacaoSalva)
	})
}
