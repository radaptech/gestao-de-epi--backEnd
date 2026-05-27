package service

import (
	"context"
	"log"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type PlanosRepository interface {
	Adicionar(ctx context.Context, arg repository.AddPlanoParams) (int32, error)
	MostrarPlanos(ctx context.Context) ([]repository.BuscaPlanosRow, error)
	AtualizarPlanos(ctx context.Context, arg repository.AtualizarPlanoParams) error
	AtualizaStatus(ctx context.Context, arg repository.AtualizarStatusPlanoParams) error
	BuscarPlanoPorNome(ctx context.Context, nome string)(repository.BuscarPlanoPorNomeRow, error)
}

type PlanosService struct {
	repo PlanosRepository
}

func NewPlanoService(p PlanosRepository) *PlanosService {

	return &PlanosService{
		repo: p,
	}
}

func (p *PlanosService) SalvarPlanos(ctx context.Context, model model.Plano) (int32, error) {

	// 1. Conversão segura de Decimal para pgtype.Numeric
	// Usar o Scan() passando a string garante que os centavos não se percam
	var mensalidadePg pgtype.Numeric
	err := mensalidadePg.Scan(model.Mensalidade.String())
	if err != nil {
		return 0, err
	}

	// 2. Função anônima (helper) para converter *int em pgtype.Int4 com segurança
	// Se for nil (ilimitado), ele devolve Valid: false, salvando NULL no banco.
	toPgInt4 := func(val *int32) pgtype.Int4 {
		if val == nil {
			return pgtype.Int4{Valid: false}
		}
		return pgtype.Int4{Int32: int32(*val), Valid: true}
	}

	// 3. Chamada ao repositório
	planoId, err := p.repo.Adicionar(ctx, repository.AddPlanoParams{
		Nome:               model.Nome,
		Mensalidade:        mensalidadePg,
		LimiteFuncionarios: toPgInt4(model.LimiteFuncionarios),
		LimiteUsuarios:     toPgInt4(model.LimiteUsuarios),
		LimiteEpis:         toPgInt4(model.LimiteEpis),
		Status:             pgtype.Text{String: model.Status, Valid: model.Status != ""},
		Descricao:          model.Descricao,
	})

	if err != nil {
		return 0, err
	}

	return planoId, nil
}

func (p *PlanosService) MostrarPlanos(ctx context.Context) ([]model.Plano, error) {

	planos, err := p.repo.MostrarPlanos(ctx)
	if err != nil {
		return []model.Plano{}, err
	}

	dto := make([]model.Plano, 0, len(planos))

	// Função auxiliar (Closure) para converter de pgtype.Int4 para *int32 respeitando o NULL
	toPtr := func(pgInt pgtype.Int4) *int32 {
		if !pgInt.Valid {
			return nil // Retorna nulo para o JSON (Ilimitado)
		}
		val := pgInt.Int32
		return &val
	}

	for _, plano := range planos {

		// Conversão exata de pgtype.Numeric para decimal.Decimal (sem passar por float)
		var valorDecimal decimal.Decimal
		if plano.Mensalidade.Valid && plano.Mensalidade.Int != nil {
			valorDecimal = decimal.NewFromBigInt(plano.Mensalidade.Int, plano.Mensalidade.Exp)
		}

		pp := model.Plano{
			ID:                 int(plano.ID),
			Nome:               plano.Nome,
			Mensalidade:        valorDecimal,
			LimiteFuncionarios: toPtr(plano.LimiteFuncionarios),
			LimiteUsuarios:     toPtr(plano.LimiteUsuarios),
			LimiteEpis:         toPtr(plano.LimiteEpis),
			Status:             plano.Status.String,
			Descricao:          plano.Descricao,
		}

		dto = append(dto, pp)
	}

	return dto, nil
}

func (p *PlanosService) AtualizarPlano(ctx context.Context, input model.AtualizarPlanoParams) error {

	// 1. Instancia a struct do repositório apenas com o ID (que é sempre obrigatório)
	params := repository.AtualizarPlanoParams{
		ID: input.ID,
	}
	// 2. Preenche os campos com segurança (só desempacota o ponteiro se ele NÃO for nil)
	if input.Nome != nil {
		params.Nome = pgtype.Text{String: *input.Nome, Valid: true}
	}

	if input.Descricao != nil {
		params.Descricao = pgtype.Text{String: *input.Descricao, Valid: true}
	}

	if input.Status != nil {
		params.Status = pgtype.Text{String: *input.Status, Valid: true}
	}

	if input.LimiteFuncionarios != nil {
		params.LimiteFuncionarios = pgtype.Int4{Int32: *input.LimiteFuncionarios, Valid: true}
	}

	if input.LimiteUsuarios != nil {
		params.LimiteUsuarios = pgtype.Int4{Int32: *input.LimiteUsuarios, Valid: true}
	}

	if input.LimiteEpis != nil {
		params.LimiteEpis = pgtype.Int4{Int32: *input.LimiteEpis, Valid: true}
	}

	if input.Mensalidade != nil {
		var numericMensalidade pgtype.Numeric

		// .Scan() lê a string gerada pelo shopspring/decimal e converte perfeitamente para o formato do banco
		err := numericMensalidade.Scan(input.Mensalidade.String())
		if err == nil {
			params.Mensalidade = numericMensalidade
		} else {
			// erro se a conversão falhar por algum motivo bizarro
			log.Printf("Erro ao converter mensalidade: %v", err)
		}
	}
	err := p.repo.AtualizarPlanos(ctx, params)

	return err
}

func (p *PlanosService) AtualizaStatus(ctx context.Context, status string, id int32)error{

	err:= p.repo.AtualizaStatus(ctx, repository.AtualizarStatusPlanoParams{
		Status: pgtype.Text{String: status, Valid: true},
		ID: id,
	})

	return err
}

