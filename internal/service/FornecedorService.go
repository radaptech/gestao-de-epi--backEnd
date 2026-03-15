package service

import (
	"context"
	"math"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

type FornecedorRepository interface {
	Adicionar(ctx context.Context, args repository.CriarFornecedorParams) error
	ListarFornecedor(ctx context.Context, args repository.ListarFornecedoresParams) ([]repository.ListarFornecedoresRow, error)
	CancelarFornecedor(ctx context.Context, args repository.DeletarFornecedorParams) (int64, error)
	AtualizaFornecedores(ctx context.Context, args repository.AtualizarFornecedorParams) (int64, error)
}

type FornecedorService struct {
	repo FornecedorRepository
}

func NewFornecedorService(repo FornecedorRepository) *FornecedorService {

	return &FornecedorService{
		repo: repo,
	}
}

func (f *FornecedorService) Adicionar(ctx context.Context, model model.FornecedorInserir, tenantId int32) error {

	model.CNPJ = strings.TrimSpace(model.CNPJ)
	model.InscricaoEstadual = strings.TrimSpace(model.InscricaoEstadual)
	model.NomeFantasia = strings.TrimSpace(model.NomeFantasia)
	model.RazaoSocial = strings.TrimSpace(model.RazaoSocial)

	err := f.repo.Adicionar(ctx, repository.CriarFornecedorParams{
		TenantID:          tenantId,
		RazaoSocial:       model.RazaoSocial,
		NomeFantasia:      model.NomeFantasia,
		Cnpj:              model.CNPJ,
		InscricaoEstadual: model.InscricaoEstadual,
	})
	if err != nil {

		return err
	}

	return nil
}

type FiltroFornecedores struct {
	Cancelados bool   `form:"cancelados"`
	Nome       string `form:"nome"`
	Cnpj       string `form:"cnpj"`
	Pagina     int32  `form:"pagina"`
	Quantidade int32  `form:"quantidade"`
}

type FornecedoresPaginados struct {
	Fornecedores []model.FornecedorDto
	Total        int64
	Pagina       int32
	PaginaFinal  int32
}

func (f *FornecedorService) ListarFornecedor(ctx context.Context, filt FiltroFornecedores, tenantId int32) (FornecedoresPaginados, error) {

	limit := filt.Quantidade
	if limit <= 0 {
		limit = 1
	}
	paginaAtual := filt.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}
	offset := max((paginaAtual-1)*limit, 0)

	filtro := repository.ListarFornecedoresParams{
		TenantID:   tenantId,
		Nome:       pgtype.Text{String: filt.Nome, Valid: filt.Nome != ""},
		Cnpj:       pgtype.Text{String: filt.Cnpj, Valid: filt.Cnpj != ""},
		Canceladas: filt.Cancelados,
		Offset:     offset,
		Limit:      limit,
	}

	fornecedores, err := f.repo.ListarFornecedor(ctx, filtro)
	if err != nil {

		return FornecedoresPaginados{}, err
	}

	dto := make([]model.FornecedorDto, 0, len(fornecedores))

	for _, fornecedor := range fornecedores {

		f := model.FornecedorDto{
			ID:                int(fornecedor.ID),
			RazaoSocial:       fornecedor.RazaoSocial,
			NomeFantasia:      fornecedor.NomeFantasia,
			CNPJ:              fornecedor.Cnpj,
			InscricaoEstadual: fornecedor.InscricaoEstadual,
		}

		dto = append(dto, f)
	}

	var total int64

	if len(fornecedores) > 0 {

		total = fornecedores[0].TotalItems
	}

	paginaFinal := int32(math.Ceil(float64(total) / float64(limit)))

	return FornecedoresPaginados{
		Fornecedores: dto,
		Total:        total,
		Pagina:       paginaAtual,
		PaginaFinal:  paginaFinal,
	}, nil
}

func (f *FornecedorService) CancelarFornecedor(ctx context.Context, id, tenantId int32) error {

	if id <= 0 {

		return helper.ErrId
	}

	arg := repository.DeletarFornecedorParams{
		ID:       id,
		TenantID: tenantId,
	}

	linhasAfetadas, err := f.repo.CancelarFornecedor(ctx, arg)
	if err != nil {

		return err
	}

	if linhasAfetadas == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil
}

func (f *FornecedorService) AtualizaFornecedor(ctx context.Context, model model.FornecedorUpdate, id, tenantId int64) error {

	if id <= 0 {

		return helper.ErrId
	}

	// 2. Prepara os dados evitando Panic em ponteiros nulos
	// Helper simples para string (pode extrair para uma função utilitária)
	toPgText := func(s *string) pgtype.Text {
		if s != nil {
			return pgtype.Text{String: *s, Valid: true}
		}
		return pgtype.Text{Valid: false} // Ou manter o valor antigo se sua query permitir COALESCE
	}

	u := repository.AtualizarFornecedorParams{
		RazaoSocial:       toPgText(model.RazaoSocial),
		NomeFantasia:      toPgText(model.NomeFantasia),
		Cnpj:              toPgText(model.CNPJ),
		InscricaoEstadual: toPgText(model.InscricaoEstadual),
		ID:                int32(id),
		TenantID:          int32(tenantId),
	}

	linhasAfetadas, err := f.repo.AtualizaFornecedores(ctx, u)
	if err != nil {

		return err
	}

	if linhasAfetadas == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil
}
