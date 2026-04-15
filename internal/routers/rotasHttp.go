package routers

import (
	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/controller"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	_ "github.com/davi-fernandesx/sistema-de-gestao-de-epi/docs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Container struct {
	Usuario      controller.LoginController
	Departamento controller.DepartamentoController
	Funcao       controller.FuncaoController
	Funcionario  controller.FuncionarioController
	Tamanho      controller.TamanhoController
	Protecao     controller.TipoProtecaoController
	Epi          controller.EpiController
	Entrada      controller.EntradaController
	Fornecedor   controller.FornecedorController
	Entrega      controller.EntregaController
	Estoque      controller.EstoqueController
}

func NewContainer(db *pgxpool.Pool) *Container {

	repoUsuario := repository.NewUsuarioRepository(db)
	repoDepartamento := repository.NewDepartamentoRepository(db)
	repoFuncao := repository.NewFuncaoRepository(db)
	repoFuncionario := repository.NewFuncionarioRepository(db)
	repoTamanho := repository.NewTamanhoRepository(db)
	repoTipoProtecao := repository.NewProtecaoRepository(db)
	repoEpi := repository.NewEpiRepository(db)
	repoEntrada := repository.NewEntradaRepository(db)
	repoFornecedor := repository.NewFornecedorRepository(db)
	repoEntrega := repository.NewEntregaRepository(db)
	repoEstoque := repository.NewEstoqueRepository(db)

	serviceUsuario := service.NewUsuarioService(repoUsuario)
	departamentoService := service.NewDepartamentoService(repoDepartamento)
	funcaoService := service.NewFuncaoService(repoFuncao)
	FornecedorService := service.NewFornecedorService(repoFornecedor)
	funcionarioService := service.NewFuncionarioService(repoFuncionario,repoEntrega ,repoEpi,db)
	tamanhoService := service.NewTamanhoService(repoTamanho)
	TipoProtecaoService := service.NewProtecaoService(repoTipoProtecao)
	epiService := service.NewEpiService(repoEpi, db)
	entradaService := service.NewEntradaService(repoEntrada, db)
	entregaService := service.NewEntregaService(repoEntrega, db)
	estoqueService := service.NewEstoqueService(repoEstoque)

	return &Container{
		Usuario:      *controller.NewLoginController(serviceUsuario),
		Departamento: *controller.NewDepartamentoController(departamentoService),
		Funcao:       *controller.NewFuncaoController(funcaoService),
		Funcionario:  *controller.NewFuncionarioController(funcionarioService),
		Tamanho:      *controller.NewTamanhoControle(tamanhoService),
		Protecao:     *controller.NewTipoProtecaoController(TipoProtecaoService),
		Epi:          *controller.NewEpiController(epiService),
		Entrada:      *controller.NewEntradaController(entradaService),
		Fornecedor:   *controller.NewFornecedorController(FornecedorService),
		Entrega:      *controller.NewEntregaController(entregaService),
		Estoque:      *controller.NewEstoqueController(estoqueService),
	}
}
func ConfigurarRotas(r *gin.Engine, c *Container, db *pgxpool.Pool) {

	queries := repository.New(db)
	// --- GRUPO 1: Rotas Públicas (Aberta) ---
	// Qualquer um acessa sem token

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"message": "Bem-vindo à API do Sistema de Gestão de EPIs - Radaptech",
			"version": "1.0.0",
		})
	})
	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"message": "API operando normalmente. Acesse a documentação do Swagger para ver as rotas.",
		})
	})
	api := r.Group("/api")
	// --- GRUPO 2: Rotas que precisam do tenentId (SaaS) ---
	// Precisa do tenant Id para passar
	api.Use(middleware.TenantMiddleware(queries))
	{

		
		api.POST("/login", c.Usuario.Login())
		api.POST("/cadastro", c.Usuario.Registrar())
	}

	// --- GRUPO 3: Rotas Protegidas (SaaS) ---
	// Precisa do Token JWT para passar
	api.Use(middleware.AutenticacaoJWT(), middleware.LoggerComUsuario())
	{

		//colaborador e adm tem acesso a essas rotas

		api.GET("/me", c.Usuario.VerPerfil())
		//departamentos
		api.GET("/departamentos", c.Departamento.ListarDepartamentos())

		//quantidade de todos os epis cadastrados
		api.GET("/quantidade-epi", c.Estoque.MostrarQuantidades())

		//valor totas dos epis
		api.GET("/saldo-epi", c.Estoque.MostrarSaldo())
		//funcao
		api.GET("/funcoes", c.Funcao.ListarFuncoes())

		//funcionario
		api.GET("/funcionarios", c.Funcionario.ListarFuncionarios())
		api.GET("/funcionario/:matricula", c.Funcionario.ListarFuncionarioPorMatricula())
		api.GET("/funcionarios-dashbord", c.Funcionario.BuscaFuncionarioDashbord())
		api.GET("/funcionario-estoque", c.Funcionario.FuncionarioCompleto())

		//tamanhos disponiveis para vincular a um epi
		api.GET("/tamanhos", c.Tamanho.ListarTodosTamanhos())
		api.GET("/tamanhos-id-epi/:id", c.Tamanho.ListarTamanhoPorIdEpi())
		api.GET("/tamanho/:id", c.Tamanho.ListarTamanhoPorId())
		

		//proteções dedicada a cada epi
		api.GET("/protecoes", c.Protecao.ListarProtecoes())
		api.GET("/protecao/:id", c.Protecao.ListarProtecaoPorId())

		//Epi´s
		api.GET("/epis", c.Epi.ListarEpis())
		api.GET("/epi/:id", c.Epi.ListarEpiPorId())
		api.GET("/epis-dashbord", c.Epi.ListarEpiDashborController())

		//entradas
		api.GET("/entradas", c.Entrada.ListarEntradas())
		api.GET("/entradas-dashbord", c.Entrada.BuscaEntradaDashbord())
		api.GET("/entradas-estoque", c.Entrada.BuscaEntradaEstoque())

		//fornecedores
		api.GET("/fornecedores", c.Fornecedor.ListarFornecedores())

		//entregas
		api.GET("/entregas", c.Entrega.ListarEntregas())
		api.POST("/cadastro-entregas", c.Entrega.Adicionar())
		api.GET("/entregas-dashbord", c.Entrega.BuscarEntregaDashbord())
		api.GET("/entrega-itens-dashbord", c.Entrega.BuscarEntregaItenDashbord())

		//rotas que apenas o "admin" tem acesso
		rotasAdm := api.Group("/gerencial")
		rotasAdm.Use(middleware.VerificaRole("admin"))
		{

			//departamentos
			rotasAdm.DELETE("/departamento/:id", c.Departamento.DeletarDepartamento())
			rotasAdm.PUT("/departamento/:id", c.Departamento.AtualizarDepartamento())
			rotasAdm.POST("/cadastro-departamento", c.Departamento.RegistraDepartamento())

			//funçoes
			rotasAdm.DELETE("/funcao/:id", c.Funcao.DeletarFuncao())
			rotasAdm.PUT("/funcao/:id", c.Funcao.AtualizarFuncao())
			rotasAdm.POST("/cadastro-funcao", c.Funcao.RegistraFuncao())

			//funcionarios
			rotasAdm.DELETE("/funcionario/:id", c.Funcionario.DeletarFuncionaioId())
			rotasAdm.PATCH("/funcionario/:id", c.Funcionario.AtualizaFuncionario())
			rotasAdm.POST("/cadastro-funcionario", c.Funcionario.Adicionar())

			//tamanhos
			rotasAdm.POST("/cadastro-tamanho", c.Tamanho.Adicionar())
			rotasAdm.DELETE("/tamanho/:id", c.Tamanho.DeletarTamanho())

			//proteções epis
			rotasAdm.POST("/cadastro-protecao", c.Protecao.AdicionarProtecao())
			rotasAdm.DELETE("/protecao/:id", c.Protecao.DeletarProtecao())

			//epis
			rotasAdm.DELETE("/epi/:id", c.Epi.DeletarEpi())
			rotasAdm.PATCH("/epi/:id", c.Epi.AtualizaEpi())
			rotasAdm.POST("/cadastro-epi", c.Epi.AdicionarEpi())

			//entrada
			rotasAdm.POST("/cadastrar-entrada", c.Entrada.AdicionarEntrada())
			rotasAdm.DELETE("/entrada/:id", c.Entrada.CancelarEntrada())

			//fornecedor
			rotasAdm.POST("/cadastro-fornecedores", c.Fornecedor.Adicionar())
			rotasAdm.DELETE("/fornecedor/:id", c.Fornecedor.CancelarFornecedor())
			rotasAdm.PATCH("/fornecedor/:id", c.Fornecedor.AtualizaFornecedor())

			//entregas
			rotasAdm.DELETE("/entrega/:id", c.Entrega.CancelarEntrega())
			rotasAdm.GET("/:matricula/ficha-pdf/:id", c.Entrega.GerarFichaEpiPDF())

			
		}
	}

}
