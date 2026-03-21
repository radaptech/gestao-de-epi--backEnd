package service

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DevolucaoRepository interface {
	AdicionarDevolucao(ctx context.Context,qtx *repository.Queries ,args repository.AddDevolucaoSimplesParams ) error
	AdicionarTroca(ctx context.Context,qtx *repository.Queries  ,arg repository.AddTrocaEpiParams) (int32, error)
	EntregaVinculada(ctx context.Context, qtx *repository.Queries ,arg repository.AddEntregaVinculadaParams) (int32, error)
	Cancelar(ctx context.Context, qtx *repository.Queries, arg repository.CancelarDevolucaoParams) (int32, error)
	Listar(ctx context.Context, args repository.ListarDevolucoesParams) ([]repository.ListarDevolucoesRow, error)
}

type DevolucaoService struct {
	repo        DevolucaoRepository
	db          *pgxpool.Pool
	queries     *repository.Queries
	repoEntrega EntregaService
}

func NewDevolucaoService(d DevolucaoRepository, db *pgxpool.Pool, repoEntregaEpi EntregaService) *DevolucaoService {

	return &DevolucaoService{

		repo:        d,
		db:          db,
		queries:     repository.New(db),
		repoEntrega: repoEntregaEpi,
	}
}

func (d *DevolucaoService) SalvarDevolucao(ctx context.Context, modelDevolucao model.DevolucaoInserir, tenantId int32) error {

	//iniciao da transação
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)
	qtx := d.queries.WithTx(tx)

	funcionario, err := d.queries.BuscaFuncionarioPorId(ctx, repository.BuscaFuncionarioPorIdParams{
		ID: int32(modelDevolucao.IdFuncionario),
		TenantID: tenantId,
	})
	if err != nil {

		return err
	}
	token := helper.GerarTokenDevolucao(funcionario.Nome,funcionario.FuncaoNome, funcionario.DepartamentoNome,modelDevolucao.DataDevolucao.Time())

	var idEpiNovo, IdTamanhoNovo, IdQuantidadeNova pgtype.Int4 //ponteiros caso item seja uma troca
	//verifica se a devolucao, tambem é uma troca
	if modelDevolucao.Troca {

		if modelDevolucao.IdEpiNovo == nil {

			return fmt.Errorf("id do novo epi é orbtigatorio")
		}

		idEpiNovo = pgtype.Int4{Int32: int32(*modelDevolucao.IdEpiNovo), Valid: true}
		IdTamanhoNovo = pgtype.Int4{Int32: int32(*modelDevolucao.IdTamanhoNovo), Valid: true}
		IdQuantidadeNova = pgtype.Int4{Int32: int32(*modelDevolucao.NovaQuantidade), Valid: true}
	}

	//verificando o motivo da devolucao
	/*caso venha com um desse 3 ids
	desgaste, dano, vencimento, o epi nao É DEVOLVIDO PARA O ESTOQUE*/
	EHDescarte := modelDevolucao.IdMotivo == 1 || modelDevolucao.IdMotivo == 2 || modelDevolucao.IdMotivo == 3

	//caso NAO SEJA UM DESCARTE
	if !EHDescarte {

		err := qtx.DevolverItemAoEstoque(ctx, repository.DevolverItemAoEstoqueParams{
			Idepi:           int32(modelDevolucao.IdEpi),
			Idtamanho:       int32(modelDevolucao.IdTamanho),
			Quantidadeatual: int32(modelDevolucao.QuantidadeADevolver),
			TenantID: tenantId,
		})
		if err != nil {
			return err
		}
	}

	arg := repository.AddTrocaEpiParams{
		TenantID: tenantId,
		Idfuncionario:         int32(modelDevolucao.IdFuncionario),
		Idepi:                 int32(modelDevolucao.IdEpi),
		Idmotivo:              int32(modelDevolucao.IdMotivo),
		DataDevolucao:         pgtype.Date{Time: modelDevolucao.DataDevolucao.Time(), Valid: true},
		Idtamanho:             int32(modelDevolucao.IdTamanho),
		Quantidadeadevolver:   int32(modelDevolucao.QuantidadeADevolver),
		Idepinovo:             idEpiNovo,
		Idtamanhonovo:         IdTamanhoNovo,
		Quantidadenova:        IdQuantidadeNova,
		AssinaturaDigital:     modelDevolucao.AssinaturaDigital,
		IDUsuarioCancelamento: pgtype.Int4{Int32: int32(modelDevolucao.IdUser), Valid: true},
		TokenValidacao: pgtype.Text{String: token, Valid: true},

	}
	/*caso o primeiro if seja falso, quer dizer que é uma devolucao simples, sem troca*/
	idDevolucao, err := d.repo.AdicionarTroca(ctx, qtx, arg) //add na tabela de devolucao
	if err != nil {
		return err
	}

	//segundo if para realização da entrega do novo epi
	if modelDevolucao.Troca {

		idtrocaConvertido := int(idDevolucao)

		modelentrega := model.EntregaParaInserir{
			ID_funcionario:     int64(arg.Idfuncionario),
			Id_user:            modelDevolucao.IdUser,
			Data_entrega:       modelDevolucao.DataDevolucao,
			IdTroca:            &idtrocaConvertido,
			Assinatura_Digital: arg.AssinaturaDigital,
			Itens: []model.ItemParaInserir{
				{
					ID_epi:     int64(*modelDevolucao.IdEpiNovo),
					ID_tamanho: int64(*modelDevolucao.IdTamanhoNovo),
					Quantidade: *modelDevolucao.NovaQuantidade,
				},
			},
		}
		err := d.repoEntrega.RegistrarEntrega(ctx, qtx, modelentrega, tenantId)
		if err != nil {

			return err
		}
	}

	return tx.Commit(ctx)
}

type FiltroDevolucao struct {
	Canceladas           bool
	EpiID                int32
	DevolucaoID          int32
	MatriculaFuncionario string
	DataInicio           configs.DataBr
	DataFim              configs.DataBr
	Pagina               int32
	Quantidade           int32
}

type DevolucaoPaginada struct {
	Devolucoes  []model.DevolucaoDto `json:"entregas"`
	Total       int64                `json:"total"`
	Pagina      int32                `json:"pagina"`
	PaginaFinal int32                `json:"pagina_final"`
}

func (d *DevolucaoService) ListarDevolucoes(ctx context.Context, f FiltroDevolucao,tenantId int32) (DevolucaoPaginada, error) {

	limit := f.Quantidade
	if limit <= 0 {
		limit = 1
	}
	paginaAtual := f.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}

	offset := max((paginaAtual-1)*limit, 0)

	filtro := repository.ListarDevolucoesParams{
		Limit:      limit,
		Offset:     offset,
		Canceladas: f.Canceladas,
		ID:         pgtype.Int4{Int32: f.DevolucaoID, Valid: f.DevolucaoID > 0},
		Matricula:  pgtype.Text{String: f.MatriculaFuncionario, Valid: f.MatriculaFuncionario != ""},
		DataInicio: pgtype.Date{Time: f.DataInicio.Time(), Valid: !f.DataInicio.IsZero()},
		DataFim:    pgtype.Date{Time: f.DataFim.Time(), Valid: !f.DataFim.IsZero()},
		TenantID: tenantId,
	}

	devolucoes, err := d.repo.Listar(ctx, filtro)
	if err != nil {
		return DevolucaoPaginada{}, err
	}

	dto := make([]model.DevolucaoDto, 0, len(devolucoes))

	for _, dev := range devolucoes {

		Matricula:= strconv.Itoa(int(dev.Matricula))
		d := model.DevolucaoDto{
			Id: int(dev.ID),
			IdFuncionario: model.Funcionario_Dto{
				ID:        int(dev.Idfuncionario),
				Nome:      dev.FuncNome,
				Matricula: Matricula,
				Funcao: model.FuncaoDto{
					ID:     int(dev.Idfuncao),
					Funcao: dev.FuncNome,
					Departamento: model.DepartamentoDto{
						ID:           int(dev.Iddepartamento),
						Departamento: dev.DepNome,
					},
				},
			},
			IdEpi: model.EpiDto{
				Id:         int(dev.Idepi),
				Nome:       dev.EpiAntigoNome,
				Fabricante: dev.EpiAntigoFab,
				CA:         dev.EpiAntigoCa,
				Tamanho: []model.TamanhoDto{
					{
						ID:      int(dev.TamAntigoID),
						Tamanho: dev.TamAntigoNome,
					},
				},
				Descricao:      dev.DescAntiga,
				DataValidadeCa: configs.DataBr(dev.ValidadeCaAntiga.Time),
				Protecao: model.TipoProtecaoDto{

					ID:   int64(dev.Idprotecaoantigo),
					Nome: dev.TipoProtecaoNomeantigo,
				},
			},
			MotivoDevolucao: model.MotivoDevolucaoEpiDto{
				Id:     int(dev.Idmotivo),
				Motivo: dev.MotivoNome,
			},
			DataDevolucao:       configs.DataBr(dev.DataDevolucao.Time),
			QuantidadeADevolver: int(dev.Quantidadeadevolver),
			AssinaturaDigital:   dev.AssinaturaDigital,
		}

		if dev.Idepinovo.Valid {

			d.IdEpiNovo = &model.EpiDto{

				Id:         int(dev.Idepinovo.Int32),
				Nome:       dev.EpiNovoNome.String,
				Fabricante: dev.EpiNovoFab.String,
				CA:         dev.EpiNovoCa.String,
				Tamanho: []model.TamanhoDto{
					{
						ID:      int(dev.Idtamanhonovo.Int32),
						Tamanho: dev.TamNovoNome.String,
					},
				},
				Descricao:      dev.DescNova.String,
				DataValidadeCa: configs.DataBr(dev.ValidadeCaNova.Time),
				Protecao: model.TipoProtecaoDto{
					ID:   int64(dev.Idprotecaonovo.Int32),
					Nome: dev.TipoProtecaoNomenovo.String,
				},
			}
		}

		dto = append(dto, d)
	}

	var total int64
	if len(devolucoes) > 0 {
		total = devolucoes[0].TotalGeral
	}

	ultimaPagina := int32(math.Ceil(float64(total) / float64(limit)))

	return DevolucaoPaginada{
		Devolucoes:  dto,
		Total:       total,
		Pagina:      paginaAtual,
		PaginaFinal: ultimaPagina,
	}, nil
}

func (d *DevolucaoService) CancelarDevolucao(ctx context.Context, id, iduser, tenatId int) error {

	if id <= 0 {
		return helper.ErrId
	}

	//abre a transaction
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	arg := repository.CancelarDevolucaoParams{
		ID:                             int32(id),
		IDUsuarioDevolucaoCancelamento: pgtype.Int4{Int32: int32(iduser), Valid: true},
		TenantID: int32(tenatId),
	}

	qtx := d.queries.WithTx(tx)
	iddevolucao, err := d.repo.Cancelar(ctx, qtx, arg) //cancela a a devolucao e me retorna seu id
	if err != nil {
		return err
	}

	//com o id da devolucao, eu cancelo a entrega, por meio do "idtroca" (caso houver uma troca nessa devolucao)
	idEntrega, err := qtx.CancelaEntregaPorIdTroca(ctx, repository.CancelaEntregaPorIdTrocaParams{
		Idtroca:                      pgtype.Int4{Int32: int32(iddevolucao), Valid: true},
		IDUsuarioEntregaCancelamento: arg.IDUsuarioDevolucaoCancelamento,
		TenantID: arg.TenantID,
	})

	if err == nil {
		/*cancelo os itens, por meio do id da entrega,
		que me retorna o id da entrada desses item e sua quantidade (por enquanto a devolucao e feita 1 para 1) */
		itensCancelados, err := qtx.CancelaItemEntregue(ctx, repository.CancelaItemEntregueParams{
			Identrega: idEntrega,
			TenantID: arg.TenantID,
		})
		if err != nil {
			return err
		}

		for _, item := range itensCancelados {

			/*agora com o id da entrada e sua quantidade, eu reponho o estoque no lote certo*/
			linhasAfetadas, err := qtx.ReporEstoqueLote(ctx, repository.ReporEstoqueLoteParams{
				Quantidadeatual: item.Quantidade,
				ID:              item.Identrada,
				TenantID: arg.TenantID,
			})
			if err != nil {
				return err
			}

			if linhasAfetadas == 0 {
				return fmt.Errorf("lote de entrada %d não encontrado para reposição", item.Identrada)
			}
		}
	} else if err != pgx.ErrNoRows {

		/*caso o erro seja diferente de ErrNorows, quer dizer que é um erro real do banco de dados, ai capturo ele*/
		return fmt.Errorf("erro ao buscar entrega de troca, %w", err)
	} else {
		return fmt.Errorf("erro ao buscar entrega de troca, %w", err)
	}

	//caso o erro seja ErrNoRows, quer dizer que foi uma troca simples e nao teve uma entrega, entao ignoramos

	return tx.Commit(ctx)
}
