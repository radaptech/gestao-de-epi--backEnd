package model

// FuncionarioInserir representa o cadastro inicial
type FuncionarioInserir struct {
	Nome            string `json:"nome" binding:"required,min=3,max=150"`
	ID_departamento int32  `json:"id_departamento" binding:"required,min=1"`
	ID_funcao       int32  `json:"id_funcao"  binding:"required,min=1"`
}

// Funcionario_Dto usado em listas e dentro de Entregas
type Funcionario_Dto struct {
	ID        int32     `json:"id"`
	Nome      string    `json:"nome"`
	Matricula string    `json:"matricula"`
	Funcao    FuncaoDto `json:"funcao"`
}

// UpdateFuncionarioRequest para atualização parcial
type UpdateFuncionarioRequest struct {
	Nome           *string `json:"nome"`
	IdDepartamento *int32  `json:"id_departamento"` 
	IdFuncao       *int32  `json:"id_funcao"`       
}

// FuncionarioDashbord para cards rápidos
type FuncionarioDashbord struct {
	Id        int32  `json:"id"`
	Nome      string `json:"nome"`
	Matricula string `json:"matricula"`
}

// FuncionarioCompletoDto para o perfil detalhado com histórico
type FuncionarioCompletoDto struct {
	ID        int32                     `json:"id"`
	Nome      string                    `json:"nome"`
	Matricula string                    `json:"matricula"`
	Funcao    FuncaoDto                 `json:"funcao"` 
	Entregas  []EntregaDoFuncionarioDto `json:"entregas"`
}