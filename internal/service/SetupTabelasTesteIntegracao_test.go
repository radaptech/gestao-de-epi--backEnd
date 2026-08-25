package service

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// criarTabelasPostgres aplica as MESMAS migrations de database/migrate que
// rodam em produção (ver configs/conn.go), em vez de manter uma cópia manual
// do schema. Isso evita que o banco de testes fique dessincronizado das
// constraints, índices e triggers reais (o que já aconteceu antes: faltavam
// vários UNIQUE por tenant e o trigger de matrícula automática).
func criarTabelasPostgres(t *testing.T, pool *pgxpool.Pool) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("não foi possível localizar o caminho das migrations")
	}
	// deste arquivo (internal/service) até a raiz do projeto: dois níveis acima
	migrationsPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "database", "migrate")

	sqlDB := stdlib.OpenDBFromPool(pool)

	driver, err := pgx.WithInstance(sqlDB, &pgx.Config{})
	if err != nil {
		t.Fatalf("erro ao iniciar driver de migração para o banco de testes: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver,
	)
	if err != nil {
		t.Fatalf("erro ao instanciar migrations no banco de testes: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("erro ao aplicar migrations no banco de testes: %v", err)
	}

	// m.Close() libera o advisory lock do Postgres e a conexão dedicada que o
	// driver do golang-migrate mantém aberta (não retorna ao pgxpool.Pool só
	// com sqlDB.Close()). Sem isso, pool.Close() no fim do teste trava para
	// sempre esperando essa conexão que nunca é devolvida.
	if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
		t.Fatalf("erro ao fechar migration (source: %v, database: %v)", srcErr, dbErr)
	}
}
