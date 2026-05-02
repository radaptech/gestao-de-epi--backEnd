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

// DevolucaoRepository define o contrato para persistência de devoluções
type DevolucaoRepository interface {
	AdicionarTroca(ctx context.Context, qtx *repository.Queries, arg repository.AddTrocaEpiParams) (int32, error)
	Cancelar(ctx context.Context, qtx *repository.Queries, arg repository.CancelarDevolucaoParams) (int32, error)
	Listar(ctx context.Context, args repository.ListarDevolucoesParams) ([]repository.ListarDevolucoesRow, error)
}

type DevolucaoService struct {
	repo        DevolucaoRepository
	db          *pgxpool.Pool
	queries     *repository.Queries
	repoEntrega EntregaService // Service de entrega para automatizar a troca
}

func NewDevolucaoService(d DevolucaoRepository, db *pgxpool.Pool, repoEntregaEpi EntregaService) *DevolucaoService {
	return &DevolucaoService{
		repo:        d,
		db:          db,
		queries:     repository.New(db),
		repoEntrega: repoEntregaEpi,
	}
}

// SalvarDevolucao orquestra a devolução de um EPI e opcionalmente realiza uma nova entrega (Troca)
func (d *DevolucaoService) SalvarDevolucao(ctx context.Context, modelDevolucao model.DevolucaoInserir, tenantId int32) error {
	// Inicia a transação atômica
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := d.queries.WithTx(tx)

	// Busca dados do funcionário para gerar o token de validade
	funcionario, err := d.queries.BuscaFuncionarioPorId(ctx, repository.BuscaFuncionarioPorIdParams{
		ID:       int32(modelDevolucao.IdFuncionario),
		TenantID: tenantId,
	})
	if err != nil {
		return fmt.Errorf("erro ao buscar funcionário: %w", err)
	}

	// Gera o token de segurança para comprovar a operação
	token := helper.GerarTokenDevolucao(funcionario.Nome, funcionario.FuncaoNome, funcionario.DepartamentoNome, modelDevolucao.DataDevolucao.Time())

	// Prepara variáveis para caso de troca (EPI novo)
	var idEpiNovo, idTamanhoNovo, idQuantidadeNova pgtype.Int4
	if modelDevolucao.Troca {
		if modelDevolucao.IdEpiNovo == nil {
			return fmt.Errorf("id do novo epi é obrigatório para trocas")
		}
		idEpiNovo = pgtype.Int4{Int32: int32(*modelDevolucao.IdEpiNovo), Valid: true}
		idTamanhoNovo = pgtype.Int4{Int32: int32(*modelDevolucao.IdTamanhoNovo), Valid: true}
		idQuantidadeNova = pgtype.Int4{Int32: int32(*modelDevolucao.NovaQuantidade), Valid: true}
	}

	// Regra de Descarte: Motivos 1 (Desgaste), 2 (Dano) ou 3 (Vencimento) não voltam ao estoque
	ehDescarte := modelDevolucao.IdMotivo == 1 || modelDevolucao.IdMotivo == 2 || modelDevolucao.IdMotivo == 3

	// Se não for descarte, devolve a quantidade ao estoque no lote mais recente
	if !ehDescarte {
		err := qtx.DevolverItemAoEstoque(ctx, repository.DevolverItemAoEstoqueParams{
			IDEpi:           int32(modelDevolucao.IdEpi),
			IDTamanho:       int32(modelDevolucao.IdTamanho),
			QuantidadeAtual: int32(modelDevolucao.QuantidadeADevolver),
			TenantID:        tenantId,
		})
		if err != nil {
			return fmt.Errorf("erro ao repor estoque: %w", err)
		}
	}

	// Registra a devolução/troca na tabela principal
	arg := repository.AddTrocaEpiParams{
		TenantID:              tenantId,
		Idfuncionario:         int32(modelDevolucao.IdFuncionario),
		Idepi:                 int32(modelDevolucao.IdEpi),
		Idmotivo:              int32(modelDevolucao.IdMotivo),
		DataDevolucao:         pgtype.Date{Time: modelDevolucao.DataDevolucao.Time(), Valid: true},
		Idtamanho:             int32(modelDevolucao.IdTamanho),
		Quantidadeadevolver:   int32(modelDevolucao.QuantidadeADevolver),
		Idepinovo:             idEpiNovo,
		Idtamanhonovo:         idTamanhoNovo,
		Quantidadenova:        idQuantidadeNova,
		AssinaturaDigital:     modelDevolucao.AssinaturaDigital,
		TokenValidacao:        pgtype.Text{String: token, Valid: true},
	}

	idDevolucao, err := d.repo.AdicionarTroca(ctx, qtx, arg)
	if err != nil {
		return fmt.Errorf("erro ao registrar devolução: %w", err)
	}

	// Se for uma troca, dispara automaticamente o processo de nova entrega
	if modelDevolucao.Troca {
		idTrocaInt := int32(idDevolucao)
		modelEntrega := model.EntregaParaInserir{
			ID_funcionario:     int32(arg.Idfuncionario),
			Data_entrega:       modelDevolucao.DataDevolucao,
			IdTroca:            &idTrocaInt,
			Assinatura_Digital: arg.AssinaturaDigital,
			Itens: []model.ItemParaInserir{
				{
					ID_epi:     int32(*modelDevolucao.IdEpiNovo),
					ID_tamanho: int32(*modelDevolucao.IdTamanhoNovo),
					Quantidade: int32(*modelDevolucao.NovaQuantidade),
				},
			},
		}
		// Chama o service de entrega dentro da mesma transação (qtx)
		if err := d.repoEntrega.RegistrarEntrega(ctx, qtx, modelEntrega, tenantId, token); err != nil {
			return fmt.Errorf("erro ao realizar nova entrega da troca: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// CancelarDevolucao reverte uma devolução, cancelando entregas vinculadas e repondo estoque
func (d *DevolucaoService) CancelarDevolucao(ctx context.Context, id, iduser, tenantId int) error {
	if id <= 0 {
		return helper.ErrId
	}

	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := d.queries.WithTx(tx)

	// Cancela o registro de devolução e obtém o ID para rastreio
	arg := repository.CancelarDevolucaoParams{
		ID:                             int32(id),
		IDUsuarioDevolucaoCancelamento: pgtype.Int4{Int32: int32(iduser), Valid: true},
		TenantID:                       int32(tenantId),
	}

	idDevolucao, err := d.repo.Cancelar(ctx, qtx, arg)
	if err != nil {
		return fmt.Errorf("falha ao cancelar devolução: %w", err)
	}

	// Tenta cancelar a entrega gerada por essa troca (se houver uma)
	idEntrega, err := qtx.CancelaEntregaPorIdTroca(ctx, repository.CancelaEntregaPorIdTrocaParams{
		Idtroca:                      pgtype.Int4{Int32: int32(idDevolucao), Valid: true},
		IDUsuarioEntregaCancelamento: arg.IDUsuarioDevolucaoCancelamento,
		TenantID:                     arg.TenantID,
	})

	// Se existia uma entrega vinculada, cancelamos os itens e devolvemos ao estoque
	if err == nil {
		itensCancelados, err := qtx.CancelaItemEntregue(ctx, repository.CancelaItemEntregueParams{
			IDEntregaCabecalho: idEntrega,
			TenantID:           arg.TenantID,
		})
		if err != nil {
			return err
		}

		for _, item := range itensCancelados {
			// Devolve a quantidade ao lote original de entrada
			linhas, err := qtx.ReporEstoqueLote(ctx, repository.ReporEstoqueLoteParams{
				QuantidadeAtual: item.Quantidade,
				ID:              item.IDEntradaItem,
				TenantID:        arg.TenantID,
			})
			if err != nil || linhas == 0 {
				return fmt.Errorf("erro ao repor estoque do lote %d", item.IDEntradaItem)
			}
		}
	} else if err != pgx.ErrNoRows {
		// Se o erro não for "sem linhas", é um erro real de banco
		return fmt.Errorf("erro crítico ao buscar entrega vinculada: %w", err)
	}

	return tx.Commit(ctx)
}

type FiltroDevolucao struct {
	Canceladas bool
	EpiID int32
	DevolucaoID int32
	MatriculaFuncionario string
	DataInicio configs.DataBr
	DataFim configs.DataBr
	Pagina int32
	Quantidade int32
}

type DevolucaoPaginada struct {
	Devolucoes []model.DevolucaoDto `json:"entregas"`
	Total int64 `json:"total"`
	Pagina int32 `json:"pagina"`
	PaginaFinal int32 `json:"pagina_final"`
}

// ListarDevolucoes gerencia a busca paginada com filtros
func (d *DevolucaoService) ListarDevolucoes(ctx context.Context, f FiltroDevolucao, tenantId int32) (DevolucaoPaginada, error) {
	limit := f.Quantidade
	if limit <= 0 {
		limit = 10
	}

	paginaAtual := f.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}
	offset := (paginaAtual - 1) * limit

	filtro := repository.ListarDevolucoesParams{
		Limit:      limit,
		Offset:     offset,
		Canceladas: f.Canceladas,
		ID:         pgtype.Int4{Int32: f.DevolucaoID, Valid: f.DevolucaoID > 0},
		Matricula:  pgtype.Text{String: f.MatriculaFuncionario, Valid: f.MatriculaFuncionario != ""},
		DataInicio: pgtype.Date{Time: f.DataInicio.Time(), Valid: !f.DataInicio.IsZero()},
		DataFim:    pgtype.Date{Time: f.DataFim.Time(), Valid: !f.DataFim.IsZero()},
		TenantID:   tenantId,
	}

	devolucoes, err := d.repo.Listar(ctx, filtro)
	if err != nil {
		return DevolucaoPaginada{}, helper.TraduzErroPostgres(err)
	}

	dto := make([]model.DevolucaoDto, 0, len(devolucoes))

	for _, dev := range devolucoes {
		matriculaStr := strconv.Itoa(int(dev.Matricula))

		// Mapeamento de dados do banco para o DTO do Frontend
		item := model.DevolucaoDto{
			Id: int(dev.ID),
			Funcionario: model.Funcionario_Dto{
				ID:        int32(dev.Idfuncionario),
				Nome:      dev.FuncNome,
				Matricula: matriculaStr,
				Funcao: model.FuncaoDto{
					ID:     int(dev.Idfuncao),
					Funcao: dev.FuncaoNome,
					Departamento: model.DepartamentoDto{
						ID:           int(dev.Iddepartamento),
						Departamento: dev.DepNome,
					},
				},
			},
			Epi: model.EpiDto{
				Id:         int32(dev.Idepi),
				Nome:       dev.EpiAntigoNome,
				Fabricante: dev.EpiAntigoFab,
				CA:         dev.EpiAntigoCa,
				Tamanhos: []model.TamanhoDto{{
					ID:      int(dev.TamAntigoID),
					Tamanho: dev.TamAntigoNome,
				}},
				Descricao:      dev.DescAntiga,
				DataValidadeCa: configs.DataBr(dev.ValidadeCaAntiga.Time),
				Protecao: model.TipoProtecaoDto{
					ID:   int64(dev.Idprotecaoantigo),
					Nome: dev.TipoProtecaoNomeantigo,
				},
			},
			Motivo: model.MotivoDevolucaoEpiDto{
				Id:     int(dev.Idmotivo),
				Motivo: dev.MotivoNome,
			},
			DataDevolucao:       configs.DataBr(dev.DataDevolucao.Time),
			QuantidadeADevolver: int(dev.Quantidadeadevolver),
			AssinaturaDigital:   dev.AssinaturaDigital,
		}

		// Se a devolução incluiu uma troca, mapeia os dados do novo EPI
		if dev.Idepinovo.Valid {
			item.EpiNovo = &model.EpiDto{
				Id:         dev.Idepinovo.Int32,
				Nome:       dev.EpiNovoNome.String,
				Fabricante: dev.EpiNovoFab.String,
				CA:         dev.EpiNovoCa.String,
				Tamanhos: []model.TamanhoDto{{
					ID:      int(dev.Idtamanhonovo.Int32),
					Tamanho: dev.TamNovoNome.String,
				}},
				Descricao:      dev.DescNova.String,
				DataValidadeCa: configs.DataBr(dev.ValidadeCaNova.Time),
				Protecao: model.TipoProtecaoDto{
					ID:   int64(dev.Idprotecaonovo.Int32),
					Nome: dev.TipoProtecaoNomenovo.String,
				},
			}
		}
		dto = append(dto, item)
	}

	var total int64
	if len(devolucoes) > 0 {
		total = devolucoes[0].TotalGeral
	}

	return DevolucaoPaginada{
		Devolucoes:  dto,
		Total:       total,
		Pagina:      paginaAtual,
		PaginaFinal: int32(math.Ceil(float64(total) / float64(limit))),
	}, nil
}
