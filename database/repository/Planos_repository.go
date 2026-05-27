package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanosRepository struct {
	q  *Queries
	db *pgxpool.Pool
}

func NewPlanosRepository(pool *pgxpool.Pool) *PlanosRepository {

	return &PlanosRepository{
		q:  New(pool),
		db: pool,
	}
}


func (p *PlanosRepository) Adicionar(ctx context.Context, arg AddPlanoParams)(int32, error){

	planoId, err:= p.q.AddPlano(ctx, arg)
	if err != nil {
		return  0, helper.TraduzErroPostgres(err)
	}

	return planoId, nil
}


func (p *PlanosRepository) MostrarPlanos(ctx context.Context)([]BuscaPlanosRow, error){

	planos,err:= p.q.BuscaPlanos(ctx)
	if err != nil {

		return []BuscaPlanosRow{}, helper.TraduzErroPostgres(err)
	}

	return planos, err
}

func (p *PlanosRepository) AtualizarPlanos(ctx context.Context, arg AtualizarPlanoParams) (error) {

	err:= p.q.AtualizarPlano(ctx, arg)
	if err != nil{

		return helper.TraduzErroPostgres(err)
	}

	return  nil

}


func (p *PlanosRepository) AtualizaStatus(ctx context.Context, arg AtualizarStatusPlanoParams)(error){

	err:= p.q.AtualizarStatusPlano(ctx, arg)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil
}

func(p *PlanosRepository) BuscarPlanoPorNome(ctx context.Context, nome string)(BuscarPlanoPorNomeRow, error){

	plano, err:= p.q.BuscarPlanoPorNome(ctx, nome)
	if err != nil {
		return BuscarPlanoPorNomeRow{}, helper.TraduzErroPostgres(err)
	}


	return plano, nil
}