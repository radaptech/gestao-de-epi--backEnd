package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
)



type ProtecaoRepository interface {

	Adicionar(ctx context.Context, nome repository.AddProtecaoParams) (repository.TipoProtecao, error)
	ListarProtecao(ctx context.Context, arg repository.BuscarProtecaoParams) (repository.BuscarProtecaoRow, error)
	ListarProtecoes(ctx context.Context, tenantId int32) ([]repository.BuscarTodasProtecoesRow, error)
	CancelarProtecao(ctx context.Context, arg repository.DeletarProtecaoParams) (int64, error)
}

type ProtecaoService struct {

	repo ProtecaoRepository
}

func NewProtecaoService(p ProtecaoRepository) *ProtecaoService {

	return &ProtecaoService{repo: p}
}

func (p *ProtecaoService) SalvarProtecao(ctx context.Context, m model.TipoProtecao, tenantId int32) (model.TipoProtecaoDto, error) {

	m.Nome = strings.TrimSpace(m.Nome)

	tp,err:= p.repo.Adicionar(ctx, repository.AddProtecaoParams{
		Nome: m.Nome,
		TenantID: tenantId,
	})
	if err != nil {
		return  model.TipoProtecaoDto{},err
	}

	return model.TipoProtecaoDto{ID: int64(tp.ID),Nome: tp.Nome}, nil
}

func (p *ProtecaoService) ListarProtecao(ctx context.Context, id int, tenatId int32) (model.TipoProtecaoDto, error){

	if id <= 0 {
		return  model.TipoProtecaoDto{},helper.ErrId
	}

	protecao, err:= p.repo.ListarProtecao(ctx, repository.BuscarProtecaoParams{
		ID: int32(id),
		TenantID: tenatId,
	})
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows){
			return model.TipoProtecaoDto{}, helper.ErrNaoEncontrado
		}
	}

	return model.TipoProtecaoDto{
		ID: int64(protecao.ID),
		Nome: protecao.Nome,
	}, nil
}

func (p *ProtecaoService) ListarProtecoes(ctx context.Context, tenantId int32) ([]model.TipoProtecaoDto, error) {

	protec, err := p.repo.ListarProtecoes(ctx, tenantId)
	if err != nil {

		return  []model.TipoProtecaoDto{}, err
	}

	protecDto := make([]model.TipoProtecaoDto, 0, len(protec))

	for _, prot := range protec {

		pro := model.TipoProtecaoDto{
			ID: int64(prot.ID),
			Nome: prot.Nome,
		}
		protecDto = append(protecDto, pro)

	}

	if protec == nil {

		return []model.TipoProtecaoDto{}, nil
	}

	return protecDto, nil

}

func (p *ProtecaoService) DeletarProtecao(ctx context.Context, id int, tenantId int32) error {


	linhas, err := p.repo.CancelarProtecao(ctx, repository.DeletarProtecaoParams{
		ID: int32(id),
		TenantID: tenantId,
	})
	if err != nil {

		return fmt.Errorf("erro ao deletar funcionario, %w, funcionario ja pode estar inativo", err)
	}

	if linhas == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil

}