package service

import (
	"context"
	"errors"

	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
)

type TamanhoRepository interface {
	Adicionar(ctx context.Context, tamanho repository.AddTamanhoParams) error
	ListarTamanho(ctx context.Context, arg repository.BuscarTamanhoParams) (repository.BuscarTamanhoRow, error)
	ListarTamanhos(ctx context.Context, tenantId int32) ([]repository.BuscarTodosTamanhosRow, error)
	CancelarTamanho(ctx context.Context, arg repository.DeletarTamanhoParams) (int64, error)
	BuscarTamanhoPorIdePI(ctx context.Context, arg repository.BuscarTamanhosPorIdEpiParams) ([]repository.BuscarTamanhosPorIdEpiRow, error)
	BuscarTamanhoPorEstoque(ctx context.Context, arg repository.BuscarTamanhosComEstoquePorEpiParams) ([]repository.BuscarTamanhosComEstoquePorEpiRow, error)
}

type TamanhoService struct {
	repo TamanhoRepository
}

func NewTamanhoService(t TamanhoRepository) *TamanhoService {

	return &TamanhoService{repo: t}
}

func (t *TamanhoService) SalvarTamanho(ctx context.Context, model model.Tamanhos, tenantId int32) error {

	model.Tamanho = strings.TrimSpace(model.Tamanho)

	if err := t.repo.Adicionar(ctx, repository.AddTamanhoParams{
		Tamanho:  model.Tamanho,
		TenantID: tenantId,
	}); err != nil {

		if errors.Is(err, helper.ErrDadoDuplicado) {

			return err
		}
	}

	return nil
}

func (t *TamanhoService) ListarTamanho(ctx context.Context, id int, tenantId int32) (model.TamanhoDto, error) {

	if id <= 0 {
		return model.TamanhoDto{}, helper.ErrId
	}

	tamanho, err := t.repo.ListarTamanho(ctx, repository.BuscarTamanhoParams{
		ID:       int32(id),
		TenantID: tenantId,
	})
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return model.TamanhoDto{}, helper.ErrNaoEncontrado
		}
		return model.TamanhoDto{}, err
	}

	return model.TamanhoDto{

		ID:      int(tamanho.ID),
		Tamanho: tamanho.Tamanho,
	}, nil
}

func (t *TamanhoService) ListarTodosTamanhos(ctx context.Context, tenantId int32) ([]model.TamanhoDto, error) {

	tamanhos, err := t.repo.ListarTamanhos(ctx, tenantId)
	if err != nil {

		return []model.TamanhoDto{}, err
	}

	tamDto := make([]model.TamanhoDto, 0, len(tamanhos))

	for _, tamanho := range tamanhos {

		tam := model.TamanhoDto{
			ID:      int(tamanho.ID),
			Tamanho: tamanho.Tamanho,
		}

		tamDto = append(tamDto, tam)
	}

	if tamDto == nil {

		return []model.TamanhoDto{}, nil
	}

	return tamDto, nil
}

func (t *TamanhoService) CancelarTamanho(ctx context.Context, id int, tenantId int32) error {

	linhas, err := t.repo.CancelarTamanho(ctx, repository.DeletarTamanhoParams{
		ID:       int32(id),
		TenantID: tenantId,
	})

	if err != nil {

		return err
	}

	if linhas == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil
}

func (t *TamanhoService) ListarTamanhoPorIdEpi(ctx context.Context, idEpi, tenantId int32) ([]model.TamanhoEntregaDto, error) {

	tamanhos, err := t.repo.BuscarTamanhoPorEstoque(ctx, repository.BuscarTamanhosComEstoquePorEpiParams{
		IDEpi: idEpi,
		TenantID: tenantId,
	})

	if err != nil {

		return []model.TamanhoEntregaDto{}, err
	}

	dto := make([]model.TamanhoEntregaDto, 0, len(tamanhos))

	for _, tamanho := range tamanhos {

		tt := model.TamanhoEntregaDto{
			ID:      int(tamanho.ID),
			Tamanho: tamanho.Tamanho,
			Id_epi:  tamanho.IDEpi,
		}

		dto = append(dto, tt)
	}

	return dto, nil

}
