package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Valida a migração 000034, que substituiu duas constraints de unicidade
// sobrepostas em `funcao` (000007: única por tenant, ignorando departamento;
// 000033: única por tenant+nome+departamento, mas sem filtro de deletado_em)
// por um único índice parcial em (tenant_id, nome, iddepartamento) WHERE
// deletado_em IS NULL.
func TestFuncaoConstraintDepartamento(t *testing.T) {

	inserirFuncao := func(db *pgxpool.Pool, tenantID, idDep int64, nome string) error {
		_, err := db.Exec(context.Background(), `
			INSERT INTO funcao (tenant_id, nome, IdDepartamento) VALUES ($1, $2, $3)
		`, tenantID, nome, idDep)
		return err
	}

	t.Run("mesmo nome em departamentos diferentes agora é permitido", func(t *testing.T) {
		db := SetupTestDB(t)
		defer db.Close()

		idPlano := CreatePlanos(t, db)
		tenantID := CreateEmpresa(t, db, idPlano)
		depFinanceiro := CreateDepartamento(t, db, tenantID)
		depRH := CreateDepartamento(t, db, tenantID)

		err := inserirFuncao(db, tenantID, depFinanceiro, "Assistente")
		require.NoError(t, err, "primeira função 'Assistente' (Financeiro) deveria ter sido criada")

		err = inserirFuncao(db, tenantID, depRH, "Assistente")
		require.NoError(t, err, "'Assistente' em um departamento diferente (RH) não deveria colidir")
	})

	t.Run("mesmo nome no mesmo departamento continua bloqueado", func(t *testing.T) {
		db := SetupTestDB(t)
		defer db.Close()

		idPlano := CreatePlanos(t, db)
		tenantID := CreateEmpresa(t, db, idPlano)
		dep := CreateDepartamento(t, db, tenantID)

		err := inserirFuncao(db, tenantID, dep, "Operador")
		require.NoError(t, err, "primeira função 'Operador' deveria ter sido criada")

		err = inserirFuncao(db, tenantID, dep, "Operador")
		require.Error(t, err, "duplicar 'Operador' no mesmo departamento deveria violar a constraint única")
		require.Contains(t, err.Error(), "idx_funcao_nome_tenant_departamento_ativo")
	})

	t.Run("soft-delete libera o nome para reuso no mesmo departamento", func(t *testing.T) {
		db := SetupTestDB(t)
		defer db.Close()

		idPlano := CreatePlanos(t, db)
		tenantID := CreateEmpresa(t, db, idPlano)
		dep := CreateDepartamento(t, db, tenantID)

		var idFuncao int64
		err := db.QueryRow(context.Background(), `
			INSERT INTO funcao (tenant_id, nome, IdDepartamento) VALUES ($1, $2, $3) RETURNING id
		`, tenantID, "Encarregado", dep).Scan(&idFuncao)
		require.NoError(t, err, "primeira função 'Encarregado' deveria ter sido criada")

		_, err = db.Exec(context.Background(), `
			UPDATE funcao SET deletado_em = now() WHERE id = $1
		`, idFuncao)
		require.NoError(t, err, "soft-delete deveria funcionar")

		err = inserirFuncao(db, tenantID, dep, "Encarregado")
		require.NoError(t, err, "após soft-delete, o nome 'Encarregado' deveria estar livre para reuso no mesmo departamento — esse era o bug corrigido pela migração 000034")
	})
}
