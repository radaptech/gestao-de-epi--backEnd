package main

import (
	"log"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/routers"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// @title           SaaS EPI API
// @version         1.0
// @description     API para gestão de EPIs.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Suporte API
// @contact.url     http://www.radaptech.com.br
// @contact.email   suporte@seusaas.com.br

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	postgressConnection := configs.ConexaoDbPostgres{}

	init := configs.Init{Conexao: &postgressConnection}

	router := gin.Default()

	// Só confia em requisições encaminhadas pela rede interna do proxy (Traefik).
	// Sem isso, o Gin confia em QUALQUER X-Forwarded-For, permitindo que o
	// cliente forje seu próprio IP e burle o rate limit por IP (ver middleware.LimitarPorIP).
	// Ajuste o CIDR se a rede de produção do proxy reverso for diferente.
	if err := router.SetTrustedProxies([]string{"172.16.0.0/12"}); err != nil {
		log.Fatal(err)
	}

	router.Use(middleware.CorsConfig(), middleware.SecurityHeaders())

	db, err := init.InitAplicattion()
	if err != nil {

		log.Fatal(err)
	}
	err = postgressConnection.RunMigrationPostgress(db)
	if err != nil {

		log.Fatal(err)
	}

	log.Println("Verificando dados iniciais...")
	err = SeedEmpresaMatriz(db)
	if err != nil {
		log.Fatalf("Falha ao criar a Empresa Matriz: %v", err)
	}

	// --- BLOCO DE REGISTRO DO VALIDATOR ---
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Aqui você registra a tag "cnpj"
		err := v.RegisterValidation("cnpj", helper.ValidateCNPJ)
		if err != nil {
			log.Fatal("Erro ao registrar validador de CNPJ")
		}
	}

	container := routers.NewContainer(db)

	routers.ConfigurarRotas(router, container, db)

	router.Run(":8080")

}
