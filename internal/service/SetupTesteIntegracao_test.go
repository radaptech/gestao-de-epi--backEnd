package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)


func SetupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	const (
		dbUser     = "postgres"
		dbPassword = "Password123!7645"
		dbName     = "testdb"
	)

	// 1. Configuração do Container Postgres
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine", // Versão leve e estável
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
			"POSTGRES_DB":       dbName,
		},
		// O Postgres é muito mais rápido que o SQL Server para subir
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2). // O Postgres loga isso duas vezes durante o boot
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Erro fatal ao subir Docker Postgres: %v", err)
	}

	// Garante que o container morra ao fim do teste
	t.Cleanup(func() {
		container.Terminate(ctx)
	})

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	// 2. Construção da String de Conexão (DSN)
	// Formato: postgres://user:password@host:port/dbname?sslmode=disable
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, host, port.Port(), dbName)

	// 3. Conectando com pgxpool (mesmo que você usa no Service)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		t.Fatalf("Erro ao configurar pool: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Erro ao abrir conexão pgx: %v", err)
	}

	// 4. Ping com retentativa
	if err := pingPostgres(ctx, pool); err != nil {
		t.Fatalf("Postgres não aceitou conexões a tempo: %v", err)
	}

	// 5. Executar migrations/scripts de criação
	// Passe o pool para criar suas tabelas
	criarTabelasPostgres(t, pool)
	
	
	return pool
}


func pingPostgres(ctx context.Context, pool *pgxpool.Pool) error {
	var err error
	for range 5 {
		err = pool.Ping(ctx)
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return err
}