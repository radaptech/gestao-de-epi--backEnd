package model

type Usuario struct {
	Nome  string `json:"nome" binding:"required,min=3,max=50"`
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,max=10"`
	Role  string `json:"cargo" binding:"required"`
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

type RecuperaLogin struct {
	Empresa  string `json:"empresa" binding:"required"`
	TenantId int    `json:"tenatId" binding:"required"`
	Email    string `json:"email" binding:"required"`
}
