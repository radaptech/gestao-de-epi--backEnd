package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EpiRepository interface {
	Adicionar(ctx context.Context, qtx *repository.Queries, epi repository.AddEpiParams) (int32, error)
	ListarEpi(ctx context.Context, arg repository.BuscarEpiParams) (repository.BuscarEpiRow, error)
	ListarEpis(ctx context.Context, args repository.BuscarTodosEpisPaginadoParams) ([]repository.BuscarTodosEpisPaginadoRow, error)
	CancelarEpi(ctx context.Context, qtx *repository.Queries, arg repository.DeletarEpiParams) (int64, error)
	AtualizaEpi(ctx context.Context, epi repository.UpdateEpiCampoParams) (int64, error)
	BuscaEpiDashbord(ctx context.Context, tenant int32) ([]repository.BuscaEpiDashbordRow, error)
	BuscaEpiTenant(ctx context.Context, tenant int32) ([]repository.BuscaTodosItensEntreguesDoTenantRow, error)
	BuscaEpiFuncionario(ctx context.Context, idfuncionario, tenantId int32)([]repository.BuscaItensEntreguesPorFuncionarioRow, error)
}

type EpiService struct {
	repo    EpiRepository
	db      *pgxpool.Pool
	queries *repository.Queries
}

func NewEpiService(repo EpiRepository, db *pgxpool.Pool) *EpiService {

	return &EpiService{
		repo:    repo,
		db:      db,
		queries: repository.New(db),
	}
}

func (e *EpiService) Salvar(ctx context.Context, model model.EpiInserir, tenantID int32) error {

	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	model.Descricao = strings.TrimSpace(model.Descricao)
	model.Fabricante = strings.TrimSpace(model.Fabricante)
	model.Nome = strings.TrimSpace(model.Nome)
	model.CA = strings.TrimSpace(model.CA)
	qtx := e.queries.WithTx(tx)

	hoje := time.Now().Truncate(24 * time.Hour)

	if model.DataValidadeCa.Time().Before(hoje) {

		return helper.ErrDataMenor
	}

	epiId, err := e.repo.Adicionar(ctx, qtx, repository.AddEpiParams{
		Nome:           model.Nome,
		Fabricante:     model.Fabricante,
		Ca:             model.CA,
		Descricao:      model.Descricao,
		ValidadeCa:     pgtype.Date{Time: model.DataValidadeCa.Time(), Valid: true},
		Idtipoprotecao: int32(model.IDProtecao),
		AlertaMinimo:   int32(model.AlertaMinimo),
		TenantID:       tenantID,
	})
	if err != nil {

		return err
	}

	for _, tamanhoId := range model.IdTamanho {
		err := qtx.AddEpiTamanho(ctx, repository.AddEpiTamanhoParams{
			Idepi:     epiId,
			Idtamanho: int32(tamanhoId),
			TenantID:  tenantID,
		})
		if err != nil {
			return helper.TraduzErroPostgres(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

type EpiFiltro struct {
	Nome       string `form:"nome"`
	Ca         string `form:"ca"`
	IdEpi      int32  `form:"idEpi"`
	Fabricante string `form:"fabricante"`
	Cancelados bool   `form:"cancelados"`
	FiltroPaginacao
}

type EpiPaginado struct {
	Epis        []model.EpiDto
	Total       int64
	Pagina      int32
	PaginaFinal int32
}

func (e *EpiService) ListarEpis(ctx context.Context, f EpiFiltro, tenantId int32) (EpiPaginado, error) {

	p := Paginacao(f.FiltroPaginacao)

	filtro := repository.BuscarTodosEpisPaginadoParams{

		Limit:      p.Limit,
		Offset:     p.Offset,
		TenantID:   tenantId,
		Nome:       pgtype.Text{String: f.Nome, Valid: f.Nome != ""},
		Ca:         pgtype.Text{String: f.Ca, Valid: f.Ca != ""},
		ID:         pgtype.Int4{Int32: f.IdEpi, Valid: f.IdEpi > 0},
		Cancelados: f.Cancelados,
		Fabricante: pgtype.Text{String: f.Fabricante, Valid: f.Fabricante != ""},
	}
	epis, err := e.repo.ListarEpis(ctx, filtro)
	if err != nil {
		return EpiPaginado{}, err
	}

	if len(epis) == 0 {
		return EpiPaginado{
			Epis:        []model.EpiDto{}, // Slice vazio
			Total:       0,
			Pagina:      0,
			PaginaFinal: 0,
		}, nil
	}

	//Possivel gargalo futuro
	todosTamanhos, err := e.queries.BuscarTodosTamanhosAgrupados(ctx, tenantId)
	if err != nil {

		return EpiPaginado{Epis: []model.EpiDto{{}}, Pagina: 0}, err
	}

	tamanhosMap := make(map[int32][]model.TamanhoDto)
	for _, t := range todosTamanhos {

		tamanhosMap[t.Idepi] = append(tamanhosMap[t.Idepi], model.TamanhoDto{
			ID:      int(t.ID),
			Tamanho: t.Tamanho,
		})
	}

	dto := make([]model.EpiDto, 0, len(epis))

	for _, epi := range epis {

		e := model.EpiDto{
			Id:             epi.ID,
			Nome:           epi.Nome,
			Fabricante:     epi.Fabricante,
			CA:             epi.Ca,
			Tamanhos:       tamanhosMap[epi.ID],
			Descricao:      epi.Descricao,
			DataValidadeCa: *configs.NewDataBrPtr(epi.ValidadeCa.Time),
			Protecao: model.TipoProtecaoDto{
				ID:   int64(epi.Idtipoprotecao),
				Nome: epi.TipoProtecaoNome,
			},
			AlertaMinimo: epi.AlertaMinimo,
		}

		if e.Tamanhos == nil {
			e.Tamanhos = []model.TamanhoDto{}
		}

		dto = append(dto, e)
	}

	var total int64
	if len(epis) > 0 {

		total = epis[0].TotalGeral

	}

	//numero da ultima pagina
	ultimaPagina := int32(math.Ceil(float64(total) / float64(p.Limit)))
	return EpiPaginado{
		Epis:        dto,
		Total:       total,
		Pagina:      p.PaginaAtual,
		PaginaFinal: ultimaPagina,
	}, nil
}

func (e *EpiService) ListarEpi(ctx context.Context, id int, tenantid int32) (model.EpiDto, error) {

	if id <= 0 {

		return model.EpiDto{}, helper.ErrId
	}

	epi, err := e.repo.ListarEpi(ctx, repository.BuscarEpiParams{
		ID:       int32(id),
		TenantID: tenantid,
	})
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {

			return model.EpiDto{}, helper.ErrNaoEncontrado
		}
		return model.EpiDto{}, err
	}

	tamanhoId, err := e.queries.BuscarTamanhosPorIdEpi(ctx, repository.BuscarTamanhosPorIdEpiParams{
		IDEpi:    epi.ID,
		TenantID: tenantid,
	})
	if err != nil {
		return model.EpiDto{}, err
	}

	tamdTO := make([]model.TamanhoDto, 0, len(tamanhoId))

	for _, tamanho := range tamanhoId {

		t := model.TamanhoDto{
			ID:      int(tamanho.ID),
			Tamanho: tamanho.Tamanho,
		}

		tamdTO = append(tamdTO, t)
	}

	return model.EpiDto{
		Id:             epi.ID,
		Nome:           epi.Nome,
		Fabricante:     epi.Fabricante,
		CA:             epi.Ca,
		Tamanhos:       tamdTO,
		Descricao:      epi.Descricao,
		DataValidadeCa: *configs.NewDataBrPtr(epi.ValidadeCa.Time),
		Protecao: model.TipoProtecaoDto{
			ID:   int64(epi.Idtipoprotecao),
			Nome: epi.TipoProtecaoNome,
		},
	}, nil
}

func (e *EpiService) CancelarEpi(ctx context.Context, id int, tenantid int32) (int64, error) {

	if id <= 0 {

		return 0, helper.ErrId
	}

	tx, err := e.db.Begin(ctx)
	if err != nil {

		return 0, err
	}

	defer tx.Rollback(ctx)
	qtx := e.queries.WithTx(tx)

	linhasAfetadas, err := e.repo.CancelarEpi(ctx, qtx, repository.DeletarEpiParams{
		ID:       int32(id),
		TenantID: tenantid,
	})
	if err != nil {
		return 0, err
	}

	if linhasAfetadas == 0 {

		return 0, helper.ErrNaoEncontrado
	}

	linhasTamanhaosId, err := qtx.DeletarTamanhosPorEpi(ctx, repository.DeletarTamanhosPorEpiParams{
		Idepi:    int32(id),
		TenantID: tenantid,
	})
	if err != nil {
		return 0, err
	}

	if linhasTamanhaosId == 0 {

		return 0, errors.New("erro de integridade: EPI ativo sem tamanhos vinculados")
	}

	if err := tx.Commit(ctx); err != nil {

		return 0, err
	}

	return linhasAfetadas, nil
}

func (e *EpiService) AtualizaEpi(ctx context.Context, model model.UpdateEpiInput, id, tenantId int32) error {

	if id <= 0 {

		return helper.ErrId
	}
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := e.queries.WithTx(tx)

	// 2. Prepara os dados evitando Panic em ponteiros nulos
	// Helper simples para string (pode extrair para uma função utilitária)
	toPgText := func(s *string) pgtype.Text {
		if s != nil {
			return pgtype.Text{String: *s, Valid: true}
		}
		return pgtype.Text{Valid: false} // Ou manter o valor antigo se sua query permitir COALESCE
	}

	// Tratamento seguro para Data
	var validadeCa pgtype.Date
	if model.DataValidadeCa != nil {

		d := model.DataValidadeCa.Time()
		dataTime := time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, time.Local) // Pega a data que veio do front
		hoje := time.Now().Truncate(24 * time.Hour)

		// Valida se a data enviada é retroativa
		if dataTime.Before(hoje) {
			return helper.ErrDataMenor
		}
		validadeCa = pgtype.Date{Time: dataTime, Valid: true}
	} else {
		validadeCa = pgtype.Date{Valid: false}
	}

	// Para Proteção
	var idProt int32
	if model.IdProtecao != nil {
		idProt = *model.IdProtecao
	}

	// Para Alerta Mínimo
	var alertaMin int32
	if model.AlertaMinimo != nil {
		alertaMin = *model.AlertaMinimo
	}

	fmt.Printf("DATA PARA O BANCO: Valid=%v, Valor=%v\n", validadeCa.Valid, validadeCa.Time)
	u := repository.UpdateEpiCampoParams{
		ID:             id,
		Nome:           toPgText(model.Nome),
		Fabricante:     toPgText(model.Fabricante),
		Ca:             toPgText(model.CA),
		Descricao:      toPgText(model.Descricao),
		ValidadeCa:     validadeCa,
		IDTipoProtecao: pgtype.Int4{Int32: idProt, Valid: model.IdProtecao != nil},
		TenantID:       tenantId,
		AlertaMinimo:   pgtype.Int4{Int32: alertaMin, Valid: model.AlertaMinimo != nil},
	}

	linhasAfetadas, err := qtx.UpdateEpiCampo(ctx, u)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	if linhasAfetadas == 0 {

		return helper.ErrNaoEncontrado
	}

	if model.Tamanhos != nil {

		_, err = qtx.DeletarTamanhosPorEpi(ctx, repository.DeletarTamanhosPorEpiParams{
			Idepi:    id,
			TenantID: tenantId,
		})
		if err != nil {

			return helper.TraduzErroPostgres(err)
		}

		for _, tamId := range model.Tamanhos {

			err := qtx.AddEpiTamanho(ctx, repository.AddEpiTamanhoParams{
				Idepi:     id,
				Idtamanho: tamId,
				TenantID:  tenantId,
			})

			if err != nil {

				return helper.TraduzErroPostgres(err)
			}
		}
	}

	return tx.Commit(ctx)
}

func (e *EpiService) ListarEpiDashbord(ctx context.Context, tenantId int32) ([]model.EpiDashBord, error) {

	epis, err := e.repo.BuscaEpiDashbord(ctx, tenantId)
	if err != nil {

		return []model.EpiDashBord{}, err
	}

	dto := make([]model.EpiDashBord, 0, len(epis))

	for _, e := range epis {

		ee := model.EpiDashBord{
			Id:           e.ID,
			Nome:         e.Nome,
			AlertaMinimo: e.AlertaMinimo,
		}

		dto = append(dto, ee)
	}

	return dto, err
}


func (e *EpiService) BuscarEpiDoFuncionario(ctx context.Context, tenantId, IdFuncionario int32) ([]model.EpiDtoDevolucao, error){


	epis, err:= e.repo.BuscaEpiFuncionario(ctx, IdFuncionario, tenantId)
	if err != nil {
		return []model.EpiDtoDevolucao{}, err
	}

	dto:= make(map[int32]*model.EpiDtoDevolucao)

	for _, i:= range epis {

		if _, ok := dto[i.IDEpi]; !ok{
			dto[i.IDEpi] = &model.EpiDtoDevolucao{
				Id: i.IDEpi,
				Nome: i.EpiNome,
				Fabricante: i.Fabricante,
				CA: i.Ca,
				Tamanhos: []model.TamanhoDto{},
				Descricao: i.Descricao,
				DataValidadeCa: configs.DataBr(i.ValidadeCa.Time),
				Protecao: model.TipoProtecaoDto{ID: int64(i.Idtipoprotecao), Nome: i.TipoProtecaoNome},
				AlertaMinimo: i.AlertaMinimo,	
				SaldoAtual: i.SaldoAtual,			
			}
		}

		dto[i.IDEpi].Tamanhos = append(dto[i.IDEpi].Tamanhos, model.TamanhoDto{
			ID: int(i.IDTamanho),
			Tamanho: i.TamanhoNome,
		} )
	}

	dtoSlice:= make([]model.EpiDtoDevolucao, 0, len(dto))
	for _, v:= range dto {

		dtoSlice = append(dtoSlice, *v)
	}

	return dtoSlice, nil
}