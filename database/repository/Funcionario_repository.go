package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
	
)



type FuncionarioRepository struct {

	q *Queries
	db *pgxpool.Pool
}

func NewFuncionarioRepository(pool *pgxpool.Pool) *FuncionarioRepository {

	return &FuncionarioRepository{
		q: New(pool),
		db: pool,
	}
}

func (f *FuncionarioRepository) Adicionar(ctx context.Context, qtx *Queries,args AddFuncionarioParams) error {

	err :=qtx.AddFuncionario(ctx, args)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil	
}

func(f *FuncionarioRepository) ListarFuncionario(ctx context.Context, arg BuscaFuncionarioParams) (BuscaFuncionarioRow, error){


	return f.q.BuscaFuncionario(ctx, arg)
}

func (f *FuncionarioRepository) ListarFuncionarios(ctx context.Context, args BuscarTodosFuncionariosParams)([]BuscarTodosFuncionariosRow, error) {

	funcs, err:= f.q.BuscarTodosFuncionarios(ctx, args)
	if err != nil {

		return []BuscarTodosFuncionariosRow{}, helper.TraduzErroPostgres(err)
	}

	return  funcs, nil
}

func (f *FuncionarioRepository) CancelarFuncionario(ctx context.Context, arg DeletarFuncionarioParams) (int64, error){

	linhasAfetadas,err:= f.q.DeletarFuncionario(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

func (f *FuncionarioRepository) AtualizarFuncionarioNome(ctx context.Context, arg UpdateFuncionarioNomeParams, qtx *Queries) (int64, error){

	linhasAfetadas,err:= qtx.UpdateFuncionarioNome(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

func (f *FuncionarioRepository) AtualizarFuncionarioDepartamento(ctx context.Context, arg UpdateFuncionarioDepartamentoParams, qtx *Queries) (int64, error){

	linhasAfetadas,err:= qtx.UpdateFuncionarioDepartamento(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

func (f *FuncionarioRepository) AtualizarFuncionarioFuncao(ctx context.Context, arg UpdateFuncionarioFuncaoParams, qtx *Queries) (int64, error){

	linhasAfetadas,err:= qtx.UpdateFuncionarioFuncao(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

func (f *FuncionarioRepository) BuscarFuncionarioDashbord(ctx context.Context, tenant int32)([]BuscaFuncionarioDashbordRow, error){

	funcionarios, err:= f.q.BuscaFuncionarioDashbord(ctx, tenant)
	if err != nil {
		return []BuscaFuncionarioDashbordRow{}, helper.TraduzErroPostgres(err)
	}

	return funcionarios, nil
}

func (f *FuncionarioRepository) BuscaFuncionarioCompleto(ctx context.Context, tenant int32)([]BuscaFuncionarioCompletoRow, error){

	funcionarios, err:= f.q.BuscaFuncionarioCompleto(ctx, tenant)
	if err != nil {

		return []BuscaFuncionarioCompletoRow{}, helper.TraduzErroPostgres(err)
	}


	return funcionarios, nil
}


func (f *FuncionarioRepository) TotalFuncionarios(ctx context.Context, qtx *Queries ,id int32) (int64, error){

	total, err:= qtx.TotalDeFuncionarios(ctx, id)
	if err != nil {

		return  0, helper.TraduzErroPostgres(err)
	}


	return total, nil
}