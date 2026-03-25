package model

type FuncionarioINserir struct {
	Nome            string `json:"nome" binding:"required,min=3,max=150"`
	ID_departamento int    `json:"id_departamento" binding:"required,min=1"`
	ID_funcao       int    `json:"id_funcao"  binding:"required,min=1"`
}

type Funcionario_Dto struct {
	ID        int       `json:"id"`
	Nome      string    `json:"nome"`
	Matricula string    `json:"matricula"`
	Funcao    FuncaoDto `json:"funcao"`
}

type UpdateFuncionarioRequest struct {
	Nome           *string `json:"nome"`            // Ponteiro! Se for nil, não atualiza
	IdDepartamento *int    `json:"id_departamento"` // Ponteiro!
	IdFuncao       *int    `json:"id_funcao"`       // Ponteiro!
}

type FuncionarioDashbord struct {

	Id int `json:"id"`
	Nome string`json:"nome"`
	Matricula string`json:"matricula"`
}
