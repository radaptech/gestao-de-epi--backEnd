package service

import (
	"context"
	"errors"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/auth"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UsuarioRepository interface {
	Cadastrar(ctx context.Context, user repository.CreateUserParams) error
	Listar(ctx context.Context, tenantId int32) ([]repository.BuscarTodosUsuariosRow, error)
	BuscarPorEmail(ctx context.Context, email repository.BuscarUsuarioPorEmailParams) (repository.BuscarUsuarioPorEmailRow, error)
	BuscarPoId(ctx context.Context, arg repository.BuscarPorIdUsuarioParams) (repository.BuscarPorIdUsuarioRow, error)
}

type UsuarioService struct {
	repo UsuarioRepository
}

func NewUsuarioService(repo UsuarioRepository) *UsuarioService {

	return &UsuarioService{repo: repo}
}

func (u *UsuarioService) Registrar(ctx context.Context, model model.Usuario, tenantId int32) error {

	model.Email = strings.TrimSpace(model.Email)
	model.Nome = strings.TrimSpace(model.Nome)
	model.Senha = strings.TrimSpace(model.Senha)
	model.Role = strings.TrimSpace(model.Role)

	novasenha, err := auth.HashPassword(model.Senha)
	if err != nil {
		return err
	}

	arg := repository.CreateUserParams{
		Nome:      model.Nome,
		Email:     model.Email,
		SenhaHash: string(novasenha),
		TenantID:  tenantId,
		Role: pgtype.Text{String: model.Role, Valid: model.Role != ""},
	}

	err = u.repo.Cadastrar(ctx, arg)
	if err != nil {

		if errors.Is(err, helper.ErrDadoDuplicado) {

			return errors.New("este email já está cadastrado")
		}

		return err
	}

	return nil
}

func (u *UsuarioService) FazerLogin(ctx context.Context, email, senha string, tenantId int32) (string, repository.BuscarUsuarioPorEmailRow, error) {

	//buscando o usuario pelo email
	usuario, err := u.repo.BuscarPorEmail(ctx, repository.BuscarUsuarioPorEmailParams{
		Email: email,
		TenantID: tenantId,
	})
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.BuscarUsuarioPorEmailRow{}, errors.New("email ou senha inválidos")
		}

		return "", repository.BuscarUsuarioPorEmailRow{}, err
	}

	//verificando se a senha bate com o hash salvo no banco de dados
	_, err = auth.HashCompare([]byte(usuario.SenhaHash), senha)
	if err != nil {

		return "", repository.BuscarUsuarioPorEmailRow{}, errors.New("email ou senha inválidos")
	}

	
	//gerando o token
	token, err := auth.GerarJWT(usuario.ID, usuario.Role.String, usuario.Nome ,tenantId)
	if err != nil {
		return "", repository.BuscarUsuarioPorEmailRow{}, errors.New("erro ao gerar token de acesso")
	}

	return token, usuario, nil
}

func (u *UsuarioService) BuscarPorId(ctx context.Context, id uint, tenantId int32) (model.RecuperaUser, error) {

	if id <= 0 {

		return model.RecuperaUser{}, helper.ErrId
	}

	usuario, err := u.repo.BuscarPoId(ctx,repository.BuscarPorIdUsuarioParams{
		ID: int32(id),
		TenantID: tenantId,
	})
	if err != nil {

		return model.RecuperaUser{}, err
	}

	return model.RecuperaUser{
		Id:    int(usuario.ID),
		Nome:  usuario.Nome,
		Email: usuario.Email,
		Role: usuario.Role.String,

	}, nil
}

func (u *UsuarioService) ListarUsuario(ctx context.Context, tenantId int32) ([]model.UsuarioResponse, error) {

	usuarios, err:= u.repo.Listar(ctx, tenantId)
	if err != nil {
		return  []model.UsuarioResponse{}, err
	}

	dto := make([]model.UsuarioResponse, 0, len(usuarios))

	for _, usuario := range usuarios {

		uu:= model.UsuarioResponse{
			ID: int(usuario.ID),
			Nome: usuario.Nome,
			Email: usuario.Email,
			Cargo: usuario.Cargo.String,
		}

		dto = append(dto, uu)
	}


	return dto, nil
}