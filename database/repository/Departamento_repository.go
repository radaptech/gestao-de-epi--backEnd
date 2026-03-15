package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)


type DepartamentoRepository struct {

	q *Queries
	db *pgxpool.Pool
}

func NewDepartamentoRepository(pool *pgxpool.Pool) *DepartamentoRepository {

	return &DepartamentoRepository{
		q: New(pool),
		db: pool,
	}
}

func (d *DepartamentoRepository) Adicionar(ctx context.Context, departamento CriaDepartamentoParams) (Departamento, error){
	
	dep,err :=d.q.CriaDepartamento(ctx, departamento)
	if err != nil {

		return Departamento{},helper.TraduzErroPostgres(err)
	}

	return dep,err
}

func (d *DepartamentoRepository) ListarDepartamento(ctx context.Context, arg BuscarDepartamentoParams) (BuscarDepartamentoRow, error){

	return d.q.BuscarDepartamento(ctx, arg)
}

func (d *DepartamentoRepository) ListarDepartamentos(ctx context.Context, args BuscarTodosDepartamentosParams)([]BuscarTodosDepartamentosRow, error){

	return  d.q.BuscarTodosDepartamentos(ctx, args)
}

func (d *DepartamentoRepository) CancelarDepartamento(ctx context.Context, arg DeletarDepartamentoParams) (int64, error){

	linhasAfetadas, err:= d.q.DeletarDepartamento(ctx, arg)

	if err != nil {

		return 0, err
	}

	return  linhasAfetadas, nil
}

func (d *DepartamentoRepository) AtualizarDepartamento(ctx context.Context, arg UpdateDepartamentoParams) (int64, error){

	return d.q.UpdateDepartamento(ctx, arg)
}