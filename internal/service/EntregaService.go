package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EntregaRepository com TODAS as funções necessárias para o sistema
type EntregaRepository interface {
	AdicionarEntrega(ctx context.Context, qtx *repository.Queries, args repository.AddEntregaEpiParams) (int32, error)
	AdicionarEntregaItem(ctx context.Context, qtx *repository.Queries, arg repository.AddItemEntregueParams) (repository.AddItemEntregueRow, error)
	ListarEntregas(ctx context.Context, args repository.ListarEntregasParams) ([]repository.ListarEntregasRow, error)
	Cancelar(ctx context.Context, qtx *repository.Queries, args repository.CancelarEntregaParams) (int32, error)
	CancelarEntregaItem(ctx context.Context, qtx *repository.Queries, arg repository.CancelaItemEntregueParams) ([]repository.CancelaItemEntregueRow, error)
	AbaterEstoqueEntrada(ctx context.Context, qtx *repository.Queries, args repository.AbaterEstoqueLoteParams) (int64, error)
	ReporEstoqueEntrada(ctx context.Context, qtx *repository.Queries, args repository.ReporEstoqueLoteParams) (int64, error)
	ListarEntregasDisponiveis(ctx context.Context, qtx *repository.Queries, args repository.ListarLotesParaConsumoParams) ([]repository.ListarLotesParaConsumoRow, error)
	ListarEpisEntreguesCancelados(ctx context.Context, qtx *repository.Queries, arg repository.ListarItensEntregueCanceladosParams) ([]repository.ListarItensEntregueCanceladosRow, error)
	ListasEntregasPorMatricula(ctx context.Context, args repository.ListarHistoricoEntregasPorMatriculaParams) ([]repository.ListarHistoricoEntregasPorMatriculaRow, error)
	BuscaEntregaDashbord(ctx context.Context, tenant int32) ([]repository.EntregaDashbordRow, error)
	BuscaEntregaItensDashbord(ctx context.Context, tenant int32) ([]repository.EntregaItensDashbordRow, error)
	BuscaTodasEntregasDoTenant(ctx context.Context, tenantId int32) ([]repository.BuscaTodasEntregasDoTenantRow, error)
}

type EntregaService struct {
	repo    EntregaRepository
	db      *pgxpool.Pool
	queries *repository.Queries
}

func NewEntregaService(r EntregaRepository, pool *pgxpool.Pool) *EntregaService {
	return &EntregaService{
		repo:    r,
		db:      pool,
		queries: repository.New(pool),
	}
}

// Salvar: Ponto de entrada para novas entregas com controle de transação
func (e *EntregaService) Salvar(ctx context.Context, model model.EntregaParaInserir, tenantid int32, token string) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		fmt.Printf("❌ [TX] Falha ao iniciar transação: %v\n", err)
		return err
	}
	defer tx.Rollback(ctx)

	qtx := e.queries.WithTx(tx)
	if err := e.RegistrarEntrega(ctx, qtx, model, tenantid, token); err != nil {
		fmt.Printf("⚠️ [SALVAR] RegistrarEntrega falhou: %v\n", err)
		return err
	}

	fmt.Println("🚀 [SALVAR] Commit realizado com sucesso!")
	return tx.Commit(ctx)
}

func (e *EntregaService) TokenEntrega(ctx context.Context, tenantId, idFuncionario int32) (string, error) {
	fmt.Printf("🔍 [TOKEN] Buscando Funcionario %d | Tenant %d\n", idFuncionario, tenantId)
	funcionario, err := e.queries.BuscaFuncionarioPorId(ctx, repository.BuscaFuncionarioPorIdParams{
		ID:       int32(idFuncionario),
		TenantID: tenantId,
	})
	if err != nil {
		fmt.Printf("❌ [TOKEN] Erro na query: %v\n", err)
		if err == pgx.ErrNoRows {
			return "", helper.ErrNaoEncontrado
		}
		return "", err
	}

	token := helper.GerarTokenAuditoria(funcionario.Nome, funcionario.FuncaoNome,
		funcionario.DepartamentoNome, time.Now())

	return token, nil
}

// RegistrarEntrega: Lógica de negócio principal (Token, Lotes e Abate de Estoque)
func (e *EntregaService) RegistrarEntrega(ctx context.Context, qtx *repository.Queries, model model.EntregaParaInserir, tenantId int32, token string) error {
	fmt.Printf("\n--- 📝 [ENTREGA] Iniciando Processo ---\n")

	var idTrocaParaBanco pgtype.Int4
	if model.IdTroca != nil {
		idTrocaParaBanco = pgtype.Int4{Int32: int32(*model.IdTroca), Valid: true}
		fmt.Printf("🔄 [TROCA] Vinculando à troca ID: %d\n", *model.IdTroca)
	}

	// 1. Salva Cabeçalho
	identrega, err := e.repo.AdicionarEntrega(ctx, qtx, repository.AddEntregaEpiParams{
		Idfuncionario:    int32(model.ID_funcionario),
		DataEntrega:      pgtype.Date{Time: model.Data_entrega.Time(), Valid: true},
		Assinatura:       model.Assinatura_Digital,
		TokenValidacao:   pgtype.Text{String: token, Valid: true},
		IDUsuarioEntrega: pgtype.Int4{Int32: int32(model.Id_user), Valid: true},
		Idtroca:          idTrocaParaBanco,
		TenantID:         tenantId,
	})
	if err != nil {
		fmt.Printf("❌ [PG] Erro no Cabeçalho: %v\n", err)
		return err
	}
	fmt.Printf("✅ [PG] Cabeçalho ID %d salvo\n", identrega)

	// 2. Loop de Itens
	for i, item := range model.Itens {
		fmt.Printf("📦 [ITEM %d] EPI %d | Tam %d | Qtd %d\n", i, item.ID_epi, item.ID_tamanho, item.Quantidade)

		quantidadeNecessaria := item.Quantidade

		lotes, err := e.repo.ListarEntregasDisponiveis(ctx, qtx, repository.ListarLotesParaConsumoParams{
			IDEpi:     int32(item.ID_epi),
			IDTamanho: int32(item.ID_tamanho),
			TenantID:  tenantId,
		})
		if err != nil {
			fmt.Printf("❌ [ESTOQUE] Erro ao buscar lotes: %v\n", err)
			return err
		}

		if len(lotes) == 0 {
			fmt.Printf("⚠️ [ESTOQUE] Saldo zerado, inativo ou vencido para EPI %d\n", item.ID_epi)
			// Essa é a mensagem que vai viajar até o React:
			return fmt.Errorf("Não foi possível entregar. O estoque deste item está zerado ou com a validade vencida.")
		}	

		for _, lote := range lotes {
			if quantidadeNecessaria <= 0 {
				break
			}

			qtdAbater := min(lote.QuantidadeAtual, int32(quantidadeNecessaria))

			// Inserir Item
			_, err := e.repo.AdicionarEntregaItem(ctx, qtx, repository.AddItemEntregueParams{
				IDEntregaCabecalho: identrega,
				IDEpi:              int32(item.ID_epi),
				IDTamanho:          int32(item.ID_tamanho),
				Quantidade:         qtdAbater,
				IDEntradaItem:      lote.ID,
				TenantID:           tenantId,
			})
			if err != nil {
				fmt.Printf("❌ [PG] Erro ao inserir item vinculado ao Lote %d: %v\n", lote.ID, err)
				return err
			}

			// Abate
			afetados, err := e.repo.AbaterEstoqueEntrada(ctx, qtx, repository.AbaterEstoqueLoteParams{
				QuantidadeAtual: qtdAbater,
				ID:              lote.ID,
				TenantID:        tenantId,
			})
			if err != nil {
				fmt.Printf("❌ [PG] Erro no abate do lote %d: %v\n", lote.ID, err)
				return err
			}
			fmt.Printf("📉 [ESTOQUE] Abatidas %d un do Lote %d (Linhas afetadas: %d)\n", qtdAbater, lote.ID, afetados)

			quantidadeNecessaria -= qtdAbater
		}

		if quantidadeNecessaria > 0 {
			return fmt.Errorf("estoque insuficiente (faltam %d unidades)", quantidadeNecessaria)
		}
	}
	return nil
}

type FiltroEntregas struct {
	Canceladas    bool           `form:"canceladas"`
	EpiID         int32          `form:"epiId"`
	EntregaID     int32          `form:"entregaId"`
	FuncionarioId int32          `form:"funcionarioId"`
	DataInicio    configs.DataBr `form:"dataInicio"`
	DataFim       configs.DataBr `form:"dataFim"`
	Pagina        int32          `form:"pagina"`
	Quantidade    int32          `form:"quantidade"`
}

type EntregaPaginada struct {
	Entradas    []model.EntregaDto `json:"entregas"`
	Total       int64              `json:"total"`
	Pagina      int32              `json:"pagina"`
	PaginaFinal int32              `json:"pagina_final"`
}

// ListaEntregas: Busca paginada para a tela principal
func (e *EntregaService) ListaEntregas(ctx context.Context, f FiltroEntregas, tenantId int32) (EntregaPaginada, error) {
	limit := f.Quantidade
	if limit <= 0 {
		limit = 10
	}
	paginaAtual := f.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}
	offset := (paginaAtual - 1) * limit

	entregas, err := e.repo.ListarEntregas(ctx, repository.ListarEntregasParams{
		Limit:         limit,
		Offset:        offset,
		Canceladas:    f.Canceladas,
		IDEntrega:     pgtype.Int4{Int32: f.EntregaID, Valid: f.EntregaID > 0},
		Idfuncionario: pgtype.Int4{Int32: f.FuncionarioId, Valid: f.FuncionarioId > 0},
		TenantID:      tenantId,
	})
	if err != nil {
		return EntregaPaginada{}, err
	}

	// Busca todos os itens para evitar N+1 queries
	todosItens, err := e.queries.BuscarTodosItensEntrega(ctx, repository.BuscarTodosItensEntregaParams{
		TenantID:  tenantId,
		IDEntrega: 0,
	})
	if err != nil {
		return EntregaPaginada{}, err
	}

	itensMap := make(map[int32][]model.ItemEntregueDto)
	for _, I := range todosItens {
		itensMap[I.EntregaID] = append(itensMap[I.EntregaID], model.ItemEntregueDto{
			Id: I.ItemID,
			Epi: model.EpiResponse{
				Id: I.EpiID, Nome: I.EpiNome, Fabricante: I.Fabricante, CA: I.Ca,
				Descricao: I.EpiDesc, DataValidadeCa: configs.DataBr(I.ValidadeCa.Time),
				Protecao: model.TipoProtecaoDto{ID: int64(I.TpID), Nome: I.TpNome},
			},
			Tamanho:    model.TamanhoDto{ID: int(I.TamID), Tamanho: I.TamNome},
			Quantidade: I.Quantidade,
		})
	}

	dto := make([]model.EntregaDto, 0, len(entregas))
	for _, ent := range entregas {
		itens := itensMap[ent.EntregaID]
		if itens == nil {
			itens = []model.ItemEntregueDto{}
		}

		dto = append(dto, model.EntregaDto{
			Id: ent.EntregaID,
			Funcionario: model.Funcionario_Dto{
				ID: int32(ent.FuncID), Nome: ent.FuncNome, Matricula: strconv.Itoa(int(ent.Matricula)),
				Funcao: model.FuncaoDto{
					ID: int(ent.FuncaoID), Funcao: ent.FuncaoNome,
					Departamento: model.DepartamentoDto{ID: int(ent.DepID), Departamento: ent.DepNome},
				},
			},
			Data_entrega:       configs.DataBr(ent.DataEntrega.Time),
			Assinatura_Digital: ent.Assinatura,
			Itens:              itens,
			Token_validacao:    ent.TokenValidacao.String,
			Id_user:            ent.IDUsuarioEntrega.Int32,
		})
	}

	var total int64
	if len(entregas) > 0 {
		total = entregas[0].TotalGeral
	}

	return EntregaPaginada{
		Entradas: dto, Total: total, Pagina: paginaAtual,
		PaginaFinal: int32(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

func (e *EntregaService) CancelarEntrega(ctx context.Context, tenantId, id, iduser int) error {
	fmt.Printf("⚠️ [CANCELAR] Iniciando cancelamento da Entrega %d\n", id)
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := e.queries.WithTx(tx)
	if err := e.RegistrarCancelamento(ctx, qtx, tenantId, id, iduser); err != nil {
		fmt.Printf("❌ [CANCELAR] Falha: %v\n", err)
		return err
	}

	fmt.Println("✅ [CANCELAR] Concluído com sucesso")
	return tx.Commit(ctx)
}

func (e *EntregaService) RegistrarCancelamento(ctx context.Context, qtx *repository.Queries, tenantID, id, iduser int) error {
	arg := repository.CancelarEntregaParams{
		ID:                           int32(id),
		IDUsuarioEntregaCancelamento: pgtype.Int4{Int32: int32(iduser), Valid: true},
		TenantID:                     int32(tenantID),
	}

	identrega, err := e.repo.Cancelar(ctx, qtx, arg)
	if err != nil || identrega == 0 {
		return helper.ErrNaoEncontrado
	}

	// Cancela itens
	_, err = e.repo.CancelarEntregaItem(ctx, qtx, repository.CancelaItemEntregueParams{
		IDEntregaCabecalho: identrega,
		TenantID:           arg.TenantID,
	})
	if err != nil {
		fmt.Printf("❌ [PG] Erro ao marcar itens como cancelados: %v\n", err)
		return err
	}

	cancelados, err := e.repo.ListarEpisEntreguesCancelados(ctx, qtx, repository.ListarItensEntregueCanceladosParams{
		IDEntregaCabecalho: identrega,
		TenantID:           arg.TenantID,
	})

	for _, c := range cancelados {
		afetados, err := e.repo.ReporEstoqueEntrada(ctx, qtx, repository.ReporEstoqueLoteParams{
			QuantidadeAtual: c.Quantidade,
			ID:              c.IDEntradaItem,
			TenantID:        arg.TenantID,
		})
		if err != nil {
			fmt.Printf("❌ [ESTOQUE] Falha ao repor Lote %d: %v\n", c.IDEntradaItem, err)
			return err
		}
		fmt.Printf("📈 [ESTOQUE] Repostas %d un no Lote %d (Afetados: %d)\n", c.Quantidade, c.IDEntradaItem, afetados)
	}
	return nil
}

// GerarDadosPdfService: Monta a estrutura para o gerador de PDF
func (e *EntregaService) GerarDadosPdfService(ctx context.Context, matricula string, idEntrega, tenantId int32) (helper.DadosPdf, error) {
	matInt, _ := strconv.Atoi(matricula)
	entregas, err := e.repo.ListasEntregasPorMatricula(ctx, repository.ListarHistoricoEntregasPorMatriculaParams{
		Matricula: int32(matInt),
		TenantID:  tenantId,
		ID:        idEntrega,
	})
	if err != nil || len(entregas) == 0 {
		return helper.DadosPdf{}, err
	}

	epis := make([]helper.DadosEpiPdf, 0)
	for _, ent := range entregas {
		epis = append(epis, helper.DadosEpiPdf{
			Data: *configs.NewDataBrPtr(ent.DataEntrega.Time), NomeEpi: ent.EpiNome,
			Ca: ent.Ca, Descricao: ent.Descricao, Tamanho: ent.Tamanho, Quantidade: ent.Quantidade,
		})
	}

	primeira := entregas[0]
	return helper.DadosPdf{
		NomeEmpresa: primeira.RazaoSocial,
		NomeFuncionario: primeira.FuncNome,
		Cpf: primeira.Cpf.String,
		Matricula: strconv.Itoa(int(primeira.Matricula)),
		Setor: primeira.DepNome,
		Cargo: primeira.FuncaoNome,
		Assinatura: primeira.Assinatura,
		Epi: epis,
	}, nil
}

// Funções de Dashboard
func (e *EntregaService) BuscaEntregaDash(ctx context.Context, tenantId int32) ([]model.EntregaDashbord, error) {
	entregas, err := e.repo.BuscaEntregaDashbord(ctx, tenantId)
	if err != nil {
		return []model.EntregaDashbord{}, err
	}

	dto := make([]model.EntregaDashbord, 0, len(entregas))
	for _, ee := range entregas {
		dto = append(dto, model.EntregaDashbord{
			Id: ee.ID, IdFuncionario: ee.Idfuncionario,
			Data_entrega: *configs.NewDataBrPtr(ee.DataEntrega.Time),
			Assinatura:   ee.Assinatura, TokenValidacao: ee.TokenValidacao.String,
		})
	}
	return dto, nil
}

func (e *EntregaService) BuscaItemDash(ctx context.Context, tenantID int32) ([]model.EntregaItensDashBord, error) {
	itens, err := e.repo.BuscaEntregaItensDashbord(ctx, tenantID)
	if err != nil {
		return []model.EntregaItensDashBord{}, err
	}

	dto := make([]model.EntregaItensDashBord, 0, len(itens))
	for _, i := range itens {
		dto = append(dto, model.EntregaItensDashBord{
			Id: i.ID, IdEntregaCabecalho: i.IDEntregaCabecalho,
			IdEpi: i.IDEpi, IdTamanho: i.IDTamanho, Quantidade: i.Quantidade,
		})
	}
	return dto, nil
}
