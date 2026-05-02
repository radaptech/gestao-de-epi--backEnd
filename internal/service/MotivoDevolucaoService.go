package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
)

type MotivoDevolucaoRepository interface {
	Adicionar(ctx context.Context, arg repository.AddMotivoDevolucaoParams) (repository.AddMotivoDevolucaoRow, error)
	ListarMotivo(ctx context.Context, arg repository.BuscaMotivoDevolucaoParams) (repository.BuscaMotivoDevolucaoRow, error)
	ListarMotivos(ctx context.Context, tenantId int32) ([]repository.BuscaTodosMotivosDevolucaoRow, error)
	CancelarMotivoDevolucao(ctx context.Context, arg repository.DeleteMotivoDevolucaoParams) (int64, error)
}

type MotivoDevolucaoService struct {
	repo MotivoDevolucaoRepository
}

func NewMotivoDevolucaoRepositoryServe(m MotivoDevolucaoRepository) *MotivoDevolucaoService {

	return &MotivoDevolucaoService{repo: m}
}

func (m *MotivoDevolucaoService) Salvar(ctx context.Context, modelM model.MotivoDevolucao, tenantId int32) (model.MotivoDevolucaoEpiDto, error) {

	modelM.Motivo = strings.TrimSpace(modelM.Motivo)

	motivo, err:= m.repo.Adicionar(ctx,repository.AddMotivoDevolucaoParams{
		TenantID: tenantId,
		Motivo: modelM.Motivo,
	})
	if err != nil {

		return model.MotivoDevolucaoEpiDto{}, err
	}

	return model.MotivoDevolucaoEpiDto{
		Id: int(motivo.ID),
		Motivo: motivo.Motivo,
	}, nil
}

func (m *MotivoDevolucaoService) ListarMotivo(ctx context.Context, id int, tenantid int32) (model.MotivoDevolucaoEpiDto, error) {

	if id <= 0 {

		return model.MotivoDevolucaoEpiDto{}, helper.ErrId
	}

	motivo, err:= m.repo.ListarMotivo(ctx,repository.BuscaMotivoDevolucaoParams{
		ID: int32(id),
		TenantID: tenantid,
	})
	if err != nil {

		return model.MotivoDevolucaoEpiDto{}, err
	}

	return model.MotivoDevolucaoEpiDto{

		Id: int(motivo.ID),
		Motivo: motivo.Motivo,
	}, nil
}

func (m *MotivoDevolucaoService) ListarMotivos(ctx context.Context, tenantId int32) ([]model.MotivoDevolucaoEpiDto, error){

	motivos, err:= m.repo.ListarMotivos(ctx, tenantId)
	if err != nil {

		return []model.MotivoDevolucaoEpiDto{},err
	}

	dto:= make([]model.MotivoDevolucaoEpiDto, 0, len(motivos))

	for _, mot := range motivos {

		M := model.MotivoDevolucaoEpiDto{
			Id: int(mot.ID),
			Motivo: mot.Motivo,
		}
		dto = append(dto, M)
	}

	return dto, nil
}

func (m *MotivoDevolucaoService) DeletarMotivo(ctx context.Context, id int, tenantId int32) error {
	
	if id <= 0 {

		return  helper.ErrId
	}

	linha,err := m.repo.CancelarMotivoDevolucao(ctx,repository.DeleteMotivoDevolucaoParams{
		ID: int32(id),
		TenantID: tenantId,
	})
	if err != nil {

		return  fmt.Errorf("erro ao deletar a funcao, %w", err)
	}


	if linha == 0 {

		return helper.ErrNaoEncontrado
	}
	
	return  nil
}