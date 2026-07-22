package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestEntrega(t *testing.T) {

	db := SetupTestDB(t)
	// Correção: Inicializar 'queries' corretamente para uso com transação manual no primeiro teste
	queries := repository.New(db)
	defer db.Close()
	ctx := context.Background()

	repo := repository.NewEntregaRepository(db)
	serv := NewEntregaService(repo, db)

	// 1. CENÁRIO SAAS: CRIAR TENANT

	planos:= CreatePlanos(t, db)
	empresa := CreateEmpresa(t, db, planos)
	user := CreateUser(t, db, empresa)
	dep := CreateDepartamento(t, db, empresa)
	funcc := CreateFuncao(t, db, dep, empresa)
	funcionario := CreateFuncionario(t, db, dep, funcc, empresa)
	protec := CreateProtecao(t, db, empresa)
	tamanho := CreateTamanho(t, db, empresa)
	epi := CreateEpi(t, db, protec, empresa)
	fornecedor := CreateFornecedor(t, db, empresa)
	entradaNf := CreateEntradaNfEpi(t, db, empresa, user,fornecedor)
	entradaEpi := CreateEntradaEpi(t, db, empresa, epi, tamanho, user, entradaNf)
	entrega := CreateEntregaEpi(t, db, funcionario, user, empresa)
	_ = CreateEpiEntregues(t, db, entrega, entradaEpi, epi, tamanho, empresa)

	entregas := []model.EntregaParaInserir{
		{
			ID_funcionario:     int32(funcionario),
			Id_user:            int32(user),
			Data_entrega:       *configs.NewDataBrPtr(time.Now()),
			Assinatura_Digital: "teste.pop",
			Itens: []model.ItemParaInserir{
				{
					ID_epi:     int32(epi),
					ID_tamanho: int32(tamanho),
					Quantidade: 10,
				},
			},
		},
	}

	t.Run("sucesso ao realizar todas as etapas de uma entrega de epi (Fluxo Manual)", func(t *testing.T) {

		tx, err := db.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx) // Segurança caso falhe antes do commit

		qtx := queries.WithTx(tx)

		// Adicionado TenantID nos parâmetros
		args := repository.AddEntregaEpiParams{
			TenantID:         int32(empresa),
			Idfuncionario:    int32(entregas[0].ID_funcionario),
			DataEntrega:      pgtype.Date{Time: entregas[0].Data_entrega.Time(), Valid: true},
			Assinatura:       entregas[0].Assinatura_Digital,
			TokenValidacao:   pgtype.Text{String: "testeToken", Valid: true},
			IDUsuarioEntrega: pgtype.Int4{Int32: int32(entregas[0].Id_user), Valid: true},
		}

		identrega, err := repo.AdicionarEntrega(ctx, qtx, args)
		require.NoError(t, err)

		for _, item := range entregas[0].Itens {

			quantidadeNescessaria := item.Quantidade

			// Adicionado TenantID nos parâmetros de busca de lote
			lotes := repository.ListarLotesParaConsumoParams{
				TenantID:  int32(empresa),
				IDEpi:     int32(item.ID_epi),
				IDTamanho: int32(item.ID_tamanho),
			}

			entradaLotes, err := repo.ListarEntregasDisponiveis(ctx, qtx, lotes)
			require.NoError(t, err)

			for _, entradaLote := range entradaLotes {

				if quantidadeNescessaria <= 0 {
					break
				}

				quantidadeAbater := min(entradaLote.QuantidadeAtual, int32(quantidadeNescessaria))

				// Adicionado TenantID ao registrar o item entregue
				itemAdd := repository.AddItemEntregueParams{
					TenantID:           int32(empresa),
					IDEntregaCabecalho: identrega,
					IDEpi:              int32(item.ID_epi),
					IDTamanho:          int32(item.ID_tamanho),
					Quantidade:         quantidadeAbater,
					IDEntradaItem:      entradaLote.ID,
				}

				// Abate no estoque (Valida tenant se sua query pedir, ou apenas ID)
				// Se sua query AbaterEstoqueLoteParams pedir tenant_id, adicione aqui.
				_, err = repo.AbaterEstoqueEntrada(ctx, qtx, repository.AbaterEstoqueLoteParams{
					QuantidadeAtual: quantidadeAbater,
					ID:              entradaLote.ID,
					TenantID:        int32(empresa), // Descomente se sua query SQL exigir
				})
				require.NoError(t, err)

				_, err = repo.AdicionarEntregaItem(ctx, qtx, itemAdd)
				require.NoError(t, err)

				quantidadeNescessaria -= quantidadeAbater

				var quantidadeAtual int32
				// Validação Manual com TenantID
				query := `select quantidade_atual from entrada_epi_item where id = $1 AND tenant_id = $2`
				err = tx.QueryRow(ctx, query, entradaLote.ID, empresa).Scan(&quantidadeAtual)
				require.NoError(t, err)

				esperado := entradaLote.QuantidadeAtual - quantidadeAbater
				require.Equal(t, esperado, quantidadeAtual)

				fmt.Printf("Lote ID: %d | Abatido: %d | Sobrou: %d\n", entradaLote.ID, quantidadeAbater, quantidadeAtual)
			}
		}

		err = tx.Commit(ctx)
		require.NoError(t, err)
	})

	t.Run("ERRO - tentar entregar mais do que tem no estoque", func(t *testing.T) {
		// Reutilizando dados do setup principal, pois são do mesmo tenant

		entregaErro := model.EntregaParaInserir{
			ID_funcionario:     int32(funcionario),
			Id_user:            int32(user),
			Data_entrega:       *configs.NewDataBrPtr(time.Now()),
			Assinatura_Digital: "teste.pop",
			IdTroca:            nil,
			Itens: []model.ItemParaInserir{
				{
					ID_epi:     int32(epi),
					ID_tamanho: int32(tamanho),
					Quantidade: 300, // Exagero intencional
				},
			},
		}

		// Passando TenantID (int32)
		err := serv.Salvar(context.Background(), entregaErro, int32(empresa), "teste123")
		require.Error(t, err)
		fmt.Println("Erro esperado recebido:", err)

		var count int

		// Validação com TenantID
		db.QueryRow(context.Background(), "SELECT count(*) FROM entrega_epi WHERE token_validacao = $1 and tenant_id = $2", "teste123", empresa).Scan(&count)
		fmt.Println(entregaErro.ID_funcionario)
		fmt.Println(empresa)
		fmt.Println("linhas afetadas: ", count)
		require.Equal(t, 0, count, "A entrega não deveria ter sido salva no banco")
	})

	 t.Run("teste de concorrencia (nao deixar 2 usuarios fazer uma entrega do mesmo lote de uma vez)", func(t *testing.T) {
		// Setup Específico para garantir isolamento deste teste
		db2 := SetupTestDB(t)
		defer db2.Close()

		repo2 := repository.NewEntregaRepository(db2)
		serv2 := NewEntregaService(repo2, db2)

		planos2:= CreatePlanos(t, db2)
		idEmpresa2 := CreateEmpresa(t, db2, planos2)
		iduser2 := CreateUser(t, db2, idEmpresa2)
		iddep2 := CreateDepartamento(t, db2, idEmpresa2)
		IdFuncao2 := CreateFuncao(t, db2, iddep2, idEmpresa2)
		idtam2 := CreateTamanho(t, db2, idEmpresa2)
		idprotec2 := CreateProtecao(t, db2, idEmpresa2)
		idepi2 := CreateEpi(t, db2, idprotec2, idEmpresa2)
		idfuncionarioC := CreateFuncionario(t, db2, iddep2, IdFuncao2, idEmpresa2)
		//fornecedores
		Idfornecedor1 := CreateFornecedor(t, db2, idEmpresa2)
		entrada2 := CreateEntradaNfEpi(t, db2, idEmpresa2, iduser2,Idfornecedor1)
		entradaEpi2 := CreateEntradaEpi1(t, db2, idEmpresa2, idepi2, idtam2, iduser2, entrada2)

		var wg sync.WaitGroup
		numRequisicoes := 2
		wg.Add(numRequisicoes)

		errs := make(chan error, numRequisicoes)

		entrega := model.EntregaParaInserir{
			ID_funcionario: int32(idfuncionarioC),
			Id_user:        int32(iduser2),
			Data_entrega:   configs.DataBr(time.Now()),
			IdTroca:        nil,
			Itens: []model.ItemParaInserir{
				{ID_epi: int32(idepi2), ID_tamanho: int32(idtam2), Quantidade: 1},
			},
		}

		for range numRequisicoes {
			go func() {
				defer wg.Done()
				// Passando TenantID
				err := serv2.Salvar(ctx, entrega, int32(idEmpresa2), "1235edgd")
				if err != nil {
					fmt.Printf("Falha na goroutine: %v\n", err)
				} else {
					fmt.Println("Sucesso na goroutine")
				}
				errs <- err
			}()
		}

		wg.Wait()
		close(errs)

		sucessos := 0
		falhas := 0

		for err := range errs {
			if err == nil {
				sucessos++
			} else {
				falhas++
			}
		}

		require.Equal(t, 1, sucessos, "Apenas uma entrega deveria ter tido sucesso")
		require.Equal(t, 1, falhas, "Uma entrega deveria ter falhado por falta de estoque")

		var qtdFinal int32
		db2.QueryRow(ctx, "SELECT quantidade_atual FROM entrada_epi_item WHERE id = $1 AND tenant_id = $2", entradaEpi2, idEmpresa2).Scan(&qtdFinal)
		require.Equal(t, int32(0), qtdFinal, "O estoque final deve ser zero e nunca negativo")
	})

	t.Run("testando sucesso ao cancelar uma entrega de epi", func(t *testing.T) {

		// 1. SETUP (Banco, Services, Helpers)
		db := SetupTestDB(t)
		defer db.Close()
		ctx := context.Background()
		planos:= CreatePlanos(t, db)
		empresa := CreateEmpresa(t, db, planos)
		repo := repository.NewEntregaRepository(db)
		serv := NewEntregaService(repo, db)

		iduser := CreateUser(t, db, empresa)
		iddep := CreateDepartamento(t, db, empresa)
		IdFuncao := CreateFuncao(t, db, iddep, empresa)
		idtam := CreateTamanho(t, db, empresa)
		idprotec := CreateProtecao(t, db, empresa)
		idepi := CreateEpi(t, db, idprotec, empresa)
		_ = CreateFuncionario(t, db, iddep, IdFuncao, empresa)
		//fornecedores
		Idfornecedor := CreateFornecedor(t, db, empresa)
		idEntrada2 := CreateEntradaNfEpi(t, db, empresa, iduser, Idfornecedor)
		idEntradaEpi2 := CreateEntradaEpi(t, db, empresa, idepi, idtam, iduser, idEntrada2)

		for i := range 4 {

			err := serv.Salvar(ctx, entregas[0], int32(empresa), "56262tdf")
			require.NoError(t, err, "A entrega %d deveria ter funcionado", i+1)
		}

		var qtdTotal int64
		db.QueryRow(ctx, "SELECT count(*) from entrada_epi_item").Scan(&qtdTotal)
		fmt.Printf("Total de entregas no banco: %d\n", qtdTotal)

		var q int64
		query := `SELECT quantidade_atual FROM entrada_epi_item WHERE id = $1`
		err := db.QueryRow(ctx, query, idEntradaEpi2).Scan(&q)
		require.NoError(t, err)

		fmt.Printf("Estoque atual do lote antes de cancelar as entregas %d: %d\n", idEntrada2, q)

		for y := range 4 {

			err := serv.CancelarEntrega(ctx, int(empresa), y+1, int(iduser))
			require.NoError(t, err, "o cancelamento %d deveria ter funcionado", y+1)

		}

		var q1 int64
		query = `SELECT quantidade_atual FROM entrada_epi_item WHERE id = $1`
		err = db.QueryRow(ctx, query, idEntradaEpi2).Scan(&q1)
		require.NoError(t, err)

		fmt.Printf("Estoque atual do lote depois de cancelar as entregas %d: %d\n", idEntrada2, q1)

	})
}
