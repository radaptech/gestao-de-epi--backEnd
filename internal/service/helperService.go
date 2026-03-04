package service

type FiltroPaginacao struct {
	Pagina     int32 `form:"pagina"`
	Quantidade int32 `form:"quantidade"`
}

type PaginacaoStruct struct {
	Limit       int32
	PaginaAtual int32
	Offset      int32
}

func Paginacao(f FiltroPaginacao) (PaginacaoStruct) {

	//quantidade de itens que o usuario que ver 
	limit := f.Quantidade
	if limit <= 0 {
		limit = 1
	}

	//limita a quantidade de itens que o usuario pode buscar
	if limit > 100 {
		limit = 100
	}
	//numero da pagina que o usuario esta tentando acessar
	paginaAtual := f.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}

	//Calcula o offset, que é o número de registros que 
	// o banco de dados precisa "ignorar"
	// ou "pular" antes de começar a retornar os resultados.
	offset := max((paginaAtual-1)*limit, 0)

	return PaginacaoStruct{
		Limit: limit,
		PaginaAtual: paginaAtual,
		Offset: offset,
	}
}
