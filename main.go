package main

import (
	"log/slog"
	"os"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/routers"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/radaptech/ginmw"

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

	slog.SetDefault(middleware.NovoLogger())

	if len(os.Args) > 1 && os.Args[1] == "backup-banco" {
		ExecutarBackupBanco(os.Args[2:])
		return
	}

	postgressConnection := configs.ConexaoDbPostgres{}

	init := configs.Init{Conexao: &postgressConnection}

	router := gin.New()
	// Vários controllers passam o *gin.Context direto pro service; sem o fallback
	// ele não enxerga o ctx do request e o slog.XxxContext perde o request_id.
	router.ContextWithFallback = true
	router.Use(middleware.LogRequest(), gin.Recovery())

	// Só confia em requisições encaminhadas pela rede interna do proxy (Traefik).
	// Sem isso, o Gin confia em QUALQUER X-Forwarded-For, permitindo que o
	// cliente forje seu próprio IP e burle o rate limit por IP (ver ginmw.RateLimit).
	// Ajuste o CIDR se a rede de produção do proxy reverso for diferente.
	if err := router.SetTrustedProxies([]string{"172.16.0.0/12"}); err != nil {
		fatal("configurar proxies confiáveis", err)
	}

	router.Use(ginmw.CORS("radaptech.com.br"), ginmw.SecurityHeaders())

	db, err := init.InitAplicattion()
	if err != nil {

		fatal("conectar no banco", err)
	}
	err = postgressConnection.RunMigrationPostgress(db)
	if err != nil {

		fatal("aplicar migrações", err)
	}

	err = SeedEmpresaMatriz(db)
	if err != nil {
		fatal("criar empresa matriz", err)
	}

	// --- BLOCO DE REGISTRO DO VALIDATOR ---
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Aqui você registra a tag "cnpj"
		err := v.RegisterValidation("cnpj", helper.ValidateCNPJ)
		if err != nil {
			fatal("registrar validador de CNPJ", err)
		}
	}

	container := routers.NewContainer(db)

	routers.ConfigurarRotas(router, container, db)

	if err := router.Run(":8080"); err != nil {
		fatal("subir servidor", err)
	}
}

// log.Fatal sairia como INFO pelo slog.SetDefault; o erro que derruba o processo é ERROR.
func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
