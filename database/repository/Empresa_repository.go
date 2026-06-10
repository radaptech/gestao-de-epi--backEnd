package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
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

func (e *EmpresaRepository) ResumoDashboard(ctx context.Context) (int64, int64, int64, int64, int64, int64, int64, float64, error) {

	var (
		EmpresasAtivas     int64
		EmpresasBloqueadas int64
		EmpresasEmTeste    int64
		TotalFuncionarios  int64
		TotalEpis          int64
		TotalEntregas      int64
		ReceitaMensal      float64
	)

	eg, gctx := errgroup.WithContext(ctx)

	eg.Go(func() error {

		var err error
		EmpresasAtivas, err = e.q.EmpresasAtivas(gctx)
		return err

	})

	eg.Go(func() error {

		var err error
		EmpresasBloqueadas, err = e.q.EmpresasBloqueadas(gctx)
		return err

	})

	eg.Go(func() error {

		var err error
		EmpresasEmTeste, err = e.q.EmpresaEmTeste(gctx)
		return err

	})

	eg.Go(func() error {

		var err error
		TotalFuncionarios, err = e.q.TotalFuncionarios(gctx)
		return err

	})

	eg.Go(func() error {

		var err error
		TotalEpis, err = e.q.TotalEpis(gctx)
		return err

	})

	eg.Go(func() error {

		var err error
		TotalEntregas, err = e.q.TotalEntregas(gctx)
		return err

	})

	eg.Go(func() error {

		var err error
		ReceitaMensal, err = e.q.ReceitaMensal(gctx)
		return err
	})

	if err:= eg.Wait(); err != nil {

		return 0,0,0,0,0,0,0,0,err
	}

	TotalEmpresas := EmpresasAtivas + EmpresasEmTeste + EmpresasBloqueadas
	return EmpresasAtivas, EmpresasBloqueadas, EmpresasEmTeste, TotalEmpresas, TotalFuncionarios, TotalEpis, TotalEntregas, ReceitaMensal, nil

}

func (e *EmpresaRepository) EmpresasRecentes(ctx context.Context) ([]EmpresasRecentesRow, error) {

	empresas, err := e.q.EmpresasRecentes(ctx)
	if err != nil {
		return []EmpresasRecentesRow{}, helper.TraduzErroPostgres(err)
	}

	return empresas, nil
}

func (e *EmpresaRepository) DadosEmpresas(ctx context.Context) ([]DadosEmpresasRow, error) {

	empresas, err := e.q.DadosEmpresas(ctx)
	if err != nil {
		return []DadosEmpresasRow{}, helper.TraduzErroPostgres(err)
	}

	return empresas, nil
}
