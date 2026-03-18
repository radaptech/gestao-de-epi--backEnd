package service

import (
	"context"
	"fmt"
	"math"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	
)

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
	ListasEntregasPorMatricula(ctx context.Context, args repository.ListarHistoricoEntregasPorMatriculaParams)([]repository.ListarHistoricoEntregasPorMatriculaRow, error)
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

func (e *EntregaService) Salvar(ctx context.Context, model model.EntregaParaInserir, tenantid int32) error {

	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := e.queries.WithTx(tx)
	err = e.RegistrarEntrega(ctx, qtx, model, tenantid)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (e *EntregaService) RegistrarEntrega(ctx context.Context, qtx *repository.Queries, model model.EntregaParaInserir, tenantId int32) error {

	funcionario, err := qtx.BuscaFuncionarioPorId(ctx, repository.BuscaFuncionarioPorIdParams{
		ID:       int32(model.ID_funcionario),
		TenantID: tenantId,
	})
	if err != nil {

		if err == pgx.ErrNoRows {

			return helper.ErrNaoEncontrado
		}
		return err
	}
	token := helper.GerarTokenAuditoria(funcionario.Nome, funcionario.FuncaoNome, funcionario.DepartamentoNome, model.Data_entrega.Time())
	// 1. Cria a variável vazia (Valid: false por padrão)
	var idTrocaParaBanco pgtype.Int4

	// 2. Verifica se veio valor. Se veio, preenche e marca como Valid: true
	if model.IdTroca != nil {
		idTrocaParaBanco = pgtype.Int4{
			Int32: int32(*model.IdTroca),
			Valid: true,
		}
	}
	args := repository.AddEntregaEpiParams{

		Idfuncionario:    int32(model.ID_funcionario),
		DataEntrega:      pgtype.Date{Time: model.Data_entrega.Time(), Valid: !model.Data_entrega.IsZero()},
		Assinatura:       model.Assinatura_Digital,
		TokenValidacao:   pgtype.Text{String: token, Valid: token != ""},
		IDUsuarioEntrega: pgtype.Int4{Int32: int32(model.Id_user), Valid: int32(model.Id_user) > 0},
		Idtroca:          idTrocaParaBanco,
		TenantID:         tenantId,
	}

	identrega, err := e.repo.AdicionarEntrega(ctx, qtx, args) //salva o "cabeçalho"
	if err != nil {

		if err == pgx.ErrNoRows {

			return helper.ErrNaoEncontrado
		}
		return err
	}

	//percorre todos os item da lista de itens
	for _, item := range model.Itens {

		quantidadeNescessaria := item.Quantidade

		lotes := repository.ListarLotesParaConsumoParams{
			Idepi:     int32(item.ID_epi),
			Idtamanho: int32(item.ID_tamanho),
			TenantID:  tenantId,
		}
		/*lista todas as entradas com quantidadeAtual maior que 0 e que tenha os idepie e idtamanhos iguais as passado nos parametros*/
		entradaLotes, err := e.repo.ListarEntregasDisponiveis(ctx, qtx, lotes)
		if err != nil {

			if err == pgx.ErrNoRows {

				return helper.ErrNaoEncontrado
			}
			return err
		}

		if len(entradaLotes) == 0 {
			return fmt.Errorf("estoque insuficiente para o EPI ID %d", item.ID_epi)
		}

		/*percorre todas as entradas achadas*/
		for _, entradaLote := range entradaLotes {

			if quantidadeNescessaria <= 0 {
				break
			}

			//escolhe o menor valor entre esses parametros
			quantidadeAbater := min(entradaLote.Quantidadeatual, int32(quantidadeNescessaria))

			itemAdd := repository.AddItemEntregueParams{
				Identrega:  identrega,
				Idepi:      int32(item.ID_epi),
				Idtamanho:  int32(item.ID_tamanho),
				Quantidade: quantidadeAbater,
				Identrada:  entradaLote.ID,
				TenantID:   tenantId,
			}

			_, err := e.repo.AdicionarEntregaItem(ctx, qtx, itemAdd)
			if err != nil {
				return err
			}

			_, err = e.repo.AbaterEstoqueEntrada(ctx, qtx, repository.AbaterEstoqueLoteParams{
				Quantidadeatual: quantidadeAbater,
				ID:              entradaLote.ID,
				TenantID:        tenantId,
			})
			if err != nil {
				return err
			}

			quantidadeNescessaria -= int(quantidadeAbater)
		}

		if quantidadeNescessaria > 0 {
			// Se sobrou quantidade, significa que percorremos todos os lotes
			// e ainda não deu o total. Rollback automático pelo defer!

			return fmt.Errorf("estoque insuficiente para o EPI ID %d (faltam %d unidades)",
				item.ID_epi, quantidadeNescessaria)
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

func (e *EntregaService) ListaEntregas(ctx context.Context, f FiltroEntregas, tenantId int32) (EntregaPaginada, error) {

	limit := f.Quantidade
	if limit <= 0 {
		limit = 1
	}
	paginaAtual := f.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}

	offset := max((paginaAtual-1)*limit, 0)

	filtro := repository.ListarEntregasParams{
		Limit:         limit,
		Offset:        offset,
		Canceladas:    f.Canceladas,
		IDEntrega:     pgtype.Int4{Int32: f.EntregaID, Valid: f.EntregaID > 0},
		IDFuncionario: pgtype.Int4{Int32: f.FuncionarioId, Valid: f.FuncionarioId > 0},
		DataInicio:    pgtype.Date{Time: f.DataInicio.Time(), Valid: !f.DataInicio.IsZero()},
		DataFim:       pgtype.Date{Time: f.DataFim.Time(), Valid: !f.DataFim.IsZero()},
		TenantID:      tenantId,
	}

	entregas, err := e.repo.ListarEntregas(ctx, filtro)
	if err != nil {

		return EntregaPaginada{}, err
	}

	todosItens, err := e.queries.BuscarTodosItensEntrega(ctx, repository.BuscarTodosItensEntregaParams{
		TenantID:  tenantId,
		IDEntrega: f.EntregaID,
	})
	if err != nil {
		return EntregaPaginada{}, err
	}

	itensMap := make(map[int32][]model.ItemEntregueDto)
	for _, I := range todosItens {

		itensMap[I.EntregaID] = append(itensMap[I.EntregaID], model.ItemEntregueDto{
			Id: int64(I.ItemID),
			Epi: model.EpiResponse{
				Id:             int(I.EpiID),
				Nome:           I.EpiNome,
				Fabricante:     I.Fabricante,
				CA:             I.Ca,
				Descricao:      I.EpiDesc,
				DataValidadeCa: configs.DataBr(I.ValidadeCa.Time),
				Protecao: model.TipoProtecaoDto{
					ID:   int64(I.TpID),
					Nome: I.TpNome,
				},
			},
			Tamanho: model.TamanhoDto{
				ID:      int(I.TamID),
				Tamanho: I.TamNome,
			},
			Quantidade: int(I.Quantidade),
		})
	}
	dto := make([]model.EntregaDto, 0, len(entregas))

	for _, entrega := range entregas {

		e := model.EntregaDto{
			Id: int64(entrega.EntregaID),
			Funcionario: model.Funcionario_Dto{
				ID:        int(entrega.FuncID),
				Nome:      entrega.FuncNome,
				Matricula: entrega.Matricula,
				Funcao: model.FuncaoDto{
					ID:     int(entrega.FuncaoID),
					Funcao: entrega.FuncaoNome,
					Departamento: model.DepartamentoDto{
						ID:           int(entrega.DepID),
						Departamento: entrega.DepNome,
					},
				},
			},
			Data_entrega:       configs.DataBr(entrega.DataEntrega.Time),
			Assinatura_Digital: entrega.Assinatura,
			Itens:              itensMap[entrega.EntregaID],
			Id_user:            int(entrega.IDUsuarioEntrega.Int32),
		}

		dto = append(dto, e)
	}

	var total int64
	if len(entregas) > 0 {
		total = entregas[0].TotalGeral
	}

	//numero da ultima pagina
	ultimaPagina := int32(math.Ceil(float64(total) / float64(limit)))

	return EntregaPaginada{
		Entradas:    dto,
		Total:       total,
		Pagina:      paginaAtual,
		PaginaFinal: ultimaPagina,
	}, nil
}

func (e *EntregaService) CancelarEntrega(ctx context.Context, tenantId, id, iduser int) error {

	if id <= 0 {

		return helper.ErrId
	}

	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := e.queries.WithTx(tx)
	err = e.RegistrarCancelamento(ctx, qtx, tenantId, id, iduser)
	if err != nil {

		return err
	}

	if err := tx.Commit(ctx); err != nil {

		return err
	}
	return nil
}

func (e *EntregaService) RegistrarCancelamento(ctx context.Context, qtx *repository.Queries, tenantID, id, iduser int) error {

	arg := repository.CancelarEntregaParams{
		ID:                           int32(id),
		IDUsuarioEntregaCancelamento: pgtype.Int4{Int32: int32(iduser), Valid: int32(iduser) > 0},
		TenantID:                     int32(tenantID),
	}

	identrega, err := e.repo.Cancelar(ctx, qtx, arg)
	if err != nil {

		return err
	}

	if identrega == 0 {

		return helper.ErrNaoEncontrado
	}

	_, err = e.repo.CancelarEntregaItem(ctx, qtx, repository.CancelaItemEntregueParams{
		Identrega: identrega,
		TenantID:  arg.TenantID,
	})
	if err != nil {

		return fmt.Errorf("erro em cancelar entrega de item, %w ", err)
	}

	cancelados, err := e.repo.ListarEpisEntreguesCancelados(ctx, qtx, repository.ListarItensEntregueCanceladosParams{
		Identrega: identrega,
		TenantID:  arg.TenantID,
	})
	if err != nil {

		return fmt.Errorf("erro em epis entregues cancelados, %w ", err)
	}

	for _, cancelado := range cancelados {

		args := repository.ReporEstoqueLoteParams{
			Quantidadeatual: cancelado.Quantidade,
			ID:              cancelado.Identrada,
			TenantID:        arg.TenantID,
		}
		linhasAfetadas, err := e.repo.ReporEstoqueEntrada(ctx, qtx, args)
		if err != nil {

			return fmt.Errorf("erro em repor estoque, %w ", err)
		}

		if linhasAfetadas == 0 {

			return fmt.Errorf("lote de entrada %d não encontrado para reposição", cancelado.Identrada)
		}

	}

	return nil
}

func (e *EntregaService) GerarDadosPdfService(ctx context.Context, matricula string, tenantId int32) (helper.DadosPdf, error) {


	entrega, err:= e.repo.ListasEntregasPorMatricula(ctx, repository.ListarHistoricoEntregasPorMatriculaParams{
		Matricula: matricula,
		TenantID: tenantId,
	})
	if err != nil {

		return helper.DadosPdf{}, err
	}

	if len(entrega) == 0 {

		return helper.DadosPdf{}, helper.ErrNaoEncontrado

	}

	epis:= make([]helper.DadosEpiPdf, 0)

	for _, ent := range entrega {

		entregaPdf:= helper.DadosEpiPdf{
			Data: *configs.NewDataBrPtr(ent.DataEntrega.Time),
			NomeEpi: ent.EpiNome,
			Ca: ent.Ca,
			Descricao: ent.Descricao,
			Tamanho: ent.Tamanho,
			Quantidade: ent.Quantidade,
		}

		epis = append(epis, entregaPdf)
	}

	pdfEntrega:= helper.DadosPdf{

		NomeEmpresa: entrega[0].RazaoSocial,
		NomeFuncionario: entrega[0].FuncNome,
		Matricula: entrega[0].Matricula,
		Setor: entrega[0].DepNome,
		Cargo: entrega[0].FuncaoNome,
		Assinatura: entrega[0].Assinatura,
		Epi: epis,
	}


	return pdfEntrega, nil
}

