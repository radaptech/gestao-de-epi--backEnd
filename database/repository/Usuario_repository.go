package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)


type UsuarioRepository struct {

	q *Queries
	db *pgxpool.Pool
}

func NewUsuarioRepository(pool *pgxpool.Pool) *UsuarioRepository {

	return &UsuarioRepository{q: New(pool), db: pool}
}

func (u *UsuarioRepository) Cadastrar(ctx context.Context, user CreateUserParams) error {

	err:= u.q.CreateUser(ctx, user)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil
}

func (u *UsuarioRepository) Listar(ctx context.Context, tenantId int32) ([]BuscarTodosUsuariosRow, error){

	id := pgtype.Int4{Int32: tenantId, Valid: true}
	usuarios, err:= u.q.BuscarTodosUsuarios(ctx, id)
	if err != nil {
		return []BuscarTodosUsuariosRow{}, helper.TraduzErroPostgres(err)
	}

	return usuarios, err
}

func(u *UsuarioRepository) BuscarPorEmail(ctx context.Context, email BuscarUsuarioPorEmailParams) (BuscarUsuarioPorEmailRow, error){

	usuario, err:= u.q.BuscarUsuarioPorEmail(ctx, email)
	if err != nil {
		return BuscarUsuarioPorEmailRow{}, helper.TraduzErroPostgres(err)
	}

	return usuario, nil
}

func (u *UsuarioRepository) BuscarPoId(ctx context.Context, arg BuscarPorIdUsuarioParams) (BuscarPorIdUsuarioRow, error){

	usuario, err:= u.q.BuscarPorIdUsuario(ctx, arg)
	if err != nil {

		return  BuscarPorIdUsuarioRow{}, helper.TraduzErroPostgres(err)
	}

	return usuario, nil
}

func(u *UsuarioRepository) RecuperaLogin(ctx context.Context, arg RecuperaLoginParams)(int32, error){

	id,err:= u.q.RecuperaLogin(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)

	}

	return id, nil
}

func (u *UsuarioRepository) SalvarToken(ctx context.Context, agr SalvarTokenRecuperacaoParams)(int64, error){

	result,err:= u.q.SalvarTokenRecuperacao(ctx, agr)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return result, nil
}


func (u *UsuarioRepository) AtualizarSenha(ctx context.Context, arg UpdateSenhaParams)(int64, error){

	result, err:= u.q.UpdateSenha(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}
 

	return result, nil
}

func (u *UsuarioRepository) RedefinirSenha(ctx context.Context, arg UpdateSenhaParams)(int64,error){

	result, err:= u.q.UpdateSenha(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return result, nil
}


func (u *UsuarioRepository) UltimoAcesso(ctx context.Context, arg AtualizarUltimoAcessoParams)(error) {

	_, err:= u.q.AtualizarUltimoAcesso(ctx, arg)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil
}

func (u *UsuarioRepository) MostrarUsuariosPainel(ctx context.Context)([]MostrarUsuariosPainelRow, error){

	usuarios, err:= u.q.MostrarUsuariosPainel(ctx)
	if err != nil {

		return []MostrarUsuariosPainelRow{}, helper.TraduzErroPostgres(err)
	}

	return usuarios, nil
}

