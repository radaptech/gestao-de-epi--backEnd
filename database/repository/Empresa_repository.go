package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmpresaRepository struct {
	q  *Queries
	db *pgxpool.Pool
}

func NewEmpresaRepository(pool *pgxpool.Pool) *EmpresaRepository {

	return &EmpresaRepository{
		q:  New(pool),
		db: pool,
	}
}

func (e *EmpresaRepository) Salvar(ctx context.Context, arg CriarEmpresaParams) error {

	err := e.q.CriarEmpresa(ctx, arg)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil
}

func (e *EmpresaRepository) ResumoDashboard(ctx context.Context) (int64, int64, int64, int64, int64, int64, int64, float64) {


	EmpresasAtivas, _ := e.q.EmpresasAtivas(ctx)
	EmpresasBloqueadas, _ := e.q.EmpresasBloqueadas(ctx)
	EmpresasEmTeste, _ := e.q.EmpresaEmTeste(ctx)
	TotalEmpresas:= EmpresasAtivas + EmpresasEmTeste + EmpresasBloqueadas
	TotalFuncionarios, _ := e.q.TotalFuncionarios(ctx)
	TotalEpis, _ := e.q.TotalEpis(ctx)
	TotalEntregas, _ := e.q.TotalEntregas(ctx)
	ReceitaMensal, _:= e.q.ReceitaMensal(ctx)


	return EmpresasAtivas, EmpresasBloqueadas,EmpresasEmTeste,TotalEmpresas, TotalFuncionarios, TotalEpis, TotalEntregas, ReceitaMensal

}


func (e *EmpresaRepository) EmpresasRecentes(ctx context.Context)([]EmpresasRecentesRow, error){


	empresas, err:= e.q.EmpresasRecentes(ctx)
	if err != nil {
		return []EmpresasRecentesRow{}, helper.TraduzErroPostgres(err)
	}


	return empresas, nil
}