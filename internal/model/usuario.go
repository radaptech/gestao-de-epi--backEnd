package model

import "time"

type Usuario struct {
	Nome         string     `json:"nome" binding:"required,min=3,max=50"`
	Email        string     `json:"email" binding:"required,email"`
	Senha        string     `json:"senha" binding:"required,max=10"`
	Role         string     `json:"role" binding:"required"`
	EmpresaID    *int       `json:"empresaId"`
	UltimoAcesso *time.Time `json:"ultimoAcesso"`
}

// LoginResponse é o que o Front vai receber
type LoginResponse struct {
	Token string `json:"token"`
	// É uma boa prática devolver dados básicos do user junto,
	// assim o front já sabe o nome sem precisar fazer outra requisição.
	User Usuario `json:"user"`
}

type LoginInput struct {
	Email string `json:"email" binding:"required,email"` // Valida formato de email
	Senha string `json:"senha" binding:"required"`       // Apenas obrigatório
}

type RecuperaUser struct {
	Id    int    `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type RecuperaUserEntrada struct {
	Id   int    `json:"id"`
	Nome string `json:"nome"`
}

type UsuarioResponse struct {
	ID    int    `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Cargo string `json:"cargo"`
}

type UsuarioResponsePainel struct {
	ID           int    `json:"id"`
	Nome         string `json:"nome"`
	Email        string `json:"email"`
	Empresa      string `json:"empresa"` // Aqui vai apenas o nome da empresa
	Tipo         string `json:"tipo"`
	Status       bool   `json:"status"`
	UltimoAcesso string `json:"ultimoAcesso"` // Formatado como string "DD/MM/YYYY HH:MM"
}

type RecuperaLogin struct {
	Empresa  string `json:"empresa" binding:"required"`
	TenantId int    `json:"-"`
	Email    string `json:"email" binding:"required"`
}

type RedefinirSenha struct {
	Token     string `json:"token" binding:"required"`
	NovaSenha string `json:"senha_nova" binding:"required,min=6"`
	TenantId  int    `json:"-"`
}

type EditarUsuarioRequest struct {
	Nome  string `json:"nome" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

type AlterarStatusRequest struct {
	Status *bool `json:"ativo" binding:"required"`
}
