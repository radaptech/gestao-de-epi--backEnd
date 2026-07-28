package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/auth"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsuarioRepository interface {
	Cadastrar(ctx context.Context, qtx *repository.Queries, user repository.CreateUserParams) error
	Listar(ctx context.Context, tenantId int32) ([]repository.BuscarTodosUsuariosRow, error)
	BuscarPorEmail(ctx context.Context, email repository.BuscarUsuarioPorEmailParams) (repository.BuscarUsuarioPorEmailRow, error)
	BuscarPoId(ctx context.Context, arg repository.BuscarPorIdUsuarioParams) (repository.BuscarPorIdUsuarioRow, error)
	RecuperaLogin(ctx context.Context, arg repository.RecuperaLoginParams) (int32, error)
	SalvarToken(ctx context.Context, agr repository.SalvarTokenRecuperacaoParams) (int64, error)
	AtualizarSenha(ctx context.Context, arg repository.UpdateSenhaParams) (int64, error)
	RedefinirSenha(ctx context.Context, arg repository.UpdateSenhaParams) (int64, error)
	UltimoAcesso(ctx context.Context, arg repository.AtualizarUltimoAcessoParams) error
	MostrarUsuariosPainel(ctx context.Context) ([]repository.MostrarUsuariosPainelRow, error)
	EditarUsuario(ctx context.Context, args repository.EditarUsuarioParams) error
	EditarStatusUsuario(ctx context.Context, arg repository.EditarStatusUsuarioParams) error
}

type UsuarioService struct {
	repo         UsuarioRepository
	emailService EmailService
	db           *pgxpool.Pool
	queries      *repository.Queries
}

func NewUsuarioService(repo UsuarioRepository, emailSrv EmailService, db *pgxpool.Pool) *UsuarioService {

	return &UsuarioService{
		repo: repo, 
		db: db,
		emailService: emailSrv,}
}

func (u *UsuarioService) Registrar(ctx context.Context, model model.Usuario) error {

	tx, err := u.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := u.queries.WithTx(tx)

	totalUsuariosPlanos, err := qtx.ValidaTotalUsuarios(ctx, int32(*model.EmpresaID))
	if err != nil {
		return err
	}
	totalUsuarioAtual, err := qtx.TotalDeUsuario(ctx, int32(*model.EmpresaID))
	if err != nil {
		return err
	}

	if totalUsuarioAtual >= int64(totalUsuariosPlanos.LimiteUsuarios.Int32) {

		return helper.ErrLimiteExcedido
	}

	model.Email = strings.TrimSpace(model.Email)
	model.Nome = strings.TrimSpace(model.Nome)
	model.Senha = strings.TrimSpace(model.Senha)
	model.Role = strings.TrimSpace(model.Role)

	novasenha, err := auth.HashPassword(model.Senha)
	if err != nil {
		return err
	}

	var tenantID int32

	// extrai o valor do ponteiro se ele NÃO for nulo
	if model.EmpresaID != nil {
		tenantID = int32(*model.EmpresaID)
	}

	arg := repository.CreateUserParams{
		Nome:      model.Nome,
		Email:     model.Email,
		SenhaHash: string(novasenha),
		TenantID:  pgtype.Int4{Int32: tenantID, Valid: model.EmpresaID != nil},
		Role:      pgtype.Text{String: model.Role, Valid: model.Role != ""},
	}

	err = u.repo.Cadastrar(ctx, qtx ,arg)
	if err != nil {

		if errors.Is(err, helper.ErrDadoDuplicado) {

			return errors.New("este email já está cadastrado")
		}

		return err
	}

	return tx.Commit(ctx)
}

func (u *UsuarioService) FazerLogin(ctx context.Context, email, senha string, tenantId int32) (string, repository.BuscarUsuarioPorEmailRow, error) {

	//buscando o usuario pelo email
	usuario, err := u.repo.BuscarPorEmail(ctx, repository.BuscarUsuarioPorEmailParams{
		Email:    email,
		TenantID: pgtype.Int4{Int32: tenantId, Valid: true},
	})
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.BuscarUsuarioPorEmailRow{}, errors.New("email ou senha inválidos")
		}

		return "", repository.BuscarUsuarioPorEmailRow{}, err
	}

	//verificando se a senha bate com o hash salvo no banco de dados
	senhaCorreta, err := auth.HashCompare([]byte(usuario.SenhaHash), senha)
	if err != nil || !senhaCorreta {

		return "", repository.BuscarUsuarioPorEmailRow{}, errors.New("email ou senha inválidos")
	}

	//gerando o token
	token, err := auth.GerarJWT(usuario.ID, usuario.Role.String, usuario.Nome, tenantId)
	if err != nil {
		return "", repository.BuscarUsuarioPorEmailRow{}, errors.New("erro ao gerar token de acesso")
	}

	return token, usuario, nil
}

func (u *UsuarioService) BuscarPorId(ctx context.Context, id uint, tenantId int32) (model.RecuperaUser, error) {

	if id <= 0 {

		return model.RecuperaUser{}, helper.ErrId
	}

	usuario, err := u.repo.BuscarPoId(ctx, repository.BuscarPorIdUsuarioParams{
		ID:       int32(id),
		TenantID: pgtype.Int4{Int32: tenantId, Valid: true},
	})
	if err != nil {

		return model.RecuperaUser{}, err
	}

	return model.RecuperaUser{
		Id:    int(usuario.ID),
		Nome:  usuario.Nome,
		Email: usuario.Email,
		Role:  usuario.Role.String,
	}, nil
}

func (u *UsuarioService) ListarUsuario(ctx context.Context, tenantId int32) ([]model.UsuarioResponse, error) {

	usuarios, err := u.repo.Listar(ctx, tenantId)
	if err != nil {
		return []model.UsuarioResponse{}, err
	}

	dto := make([]model.UsuarioResponse, 0, len(usuarios))

	for _, usuario := range usuarios {

		uu := model.UsuarioResponse{
			ID:    int(usuario.ID),
			Nome:  usuario.Nome,
			Email: usuario.Email,
			Cargo: usuario.Cargo.String,
		}

		dto = append(dto, uu)
	}

	return dto, nil
}

func (u *UsuarioService) RecuperacaoSenha(ctx context.Context, rl model.RecuperaLogin) error {

	//verifica se o email existe
	id, err := u.repo.RecuperaLogin(ctx, repository.RecuperaLoginParams{
		Email:    rl.Email,
		TenantID: pgtype.Int4{Int32: int32(rl.TenantId), Valid: true},
	})
	if err != nil {

		return err
	}

	//apos isso, gera o token
	token := uuid.New().String()
	tempoExpiracao := time.Now().UTC().Add(10 * time.Minute)

	_, err = u.repo.SalvarToken(ctx, repository.SalvarTokenRecuperacaoParams{
		TokenRecuperacaoSenha: pgtype.Text{String: token, Valid: token != ""},
		TokenExpiracao:        pgtype.Timestamp{Time: tempoExpiracao, Valid: true},
		ID:                    id,
		TenantID:              pgtype.Int4{Int32: int32(rl.TenantId), Valid: true},
	})
	if err != nil {

		return err
	}

	err = u.emailService.EnviarEmailRecuperacao(rl.Email, token, rl.Empresa)
	if err != nil {

		return fmt.Errorf("erro ao enviar email: %w", err)
	}
	return nil

}

func (u *UsuarioService) RedefinirSenha(ctx context.Context, rs model.RedefinirSenha) error {

	senhaByte, err := auth.HashPassword(rs.NovaSenha)
	if err != nil {
		log.Printf("erro ao gerar hars da senha: %v", err)
		return fmt.Errorf("erro ao gerar hash da senha")
	}

	linha, err := u.repo.RedefinirSenha(ctx, repository.UpdateSenhaParams{
		SenhaHash:             string(senhaByte),
		TokenRecuperacaoSenha: pgtype.Text{String: rs.Token, Valid: true},
		TenantID:              pgtype.Int4{Int32: int32(rs.TenantId), Valid: true},
	})

	if linha == 0 {

		return errors.New("link inválido, expirado ou não pertence a esta empresa")
	}

	return nil
}

func (u *UsuarioService) UltimoAcesso(ctx context.Context, id, tenantId int32) error {

	err := u.repo.UltimoAcesso(ctx, repository.AtualizarUltimoAcessoParams{
		ID:       id,
		TenantID: pgtype.Int4{Int32: tenantId, Valid: true},
	})
	if err != nil {

		return err
	}

	return nil
}

func (u *UsuarioService) MostrarUsuariosPainel(ctx context.Context) ([]model.UsuarioResponsePainel, error) {

	usuarios, err := u.repo.MostrarUsuariosPainel(ctx)
	if err != nil {

		return []model.UsuarioResponsePainel{}, err
	}

	dto := make([]model.UsuarioResponsePainel, 0, len(usuarios))

	for _, usuario := range usuarios {

		uu := model.UsuarioResponsePainel{
			ID:           int(usuario.ID),
			Nome:         usuario.Nome,
			Email:        usuario.Email,
			Empresa:      usuario.Empresa,
			Tipo:         usuario.Tipo.String,
			Status:       usuario.Ativo.Bool,
			UltimoAcesso: usuario.UltimoAcesso.Time.Format("02/01/2006 15:04"),
		}

		dto = append(dto, uu)
	}

	return dto, nil
}

func (u *UsuarioService) EditarUsuario(ctx context.Context, id int32, model model.EditarUsuarioRequest) error {

	err := u.repo.EditarUsuario(ctx, repository.EditarUsuarioParams{
		Nome:  model.Nome,
		Email: model.Email,
		Role:  pgtype.Text{String: model.Role, Valid: model.Role != ""},
		ID:    id,
	})

	if err != nil {

		return err
	}

	return nil
}

func (u *UsuarioService) EditarStatusUsuario(ctx context.Context, id int32, model model.AlterarStatusRequest) error {

	err := u.repo.EditarStatusUsuario(ctx, repository.EditarStatusUsuarioParams{
		Ativo: pgtype.Bool{Bool: *model.Status, Valid: true},
		ID:    id,
	})
	if err != nil {

		return err
	}

	return nil
}
