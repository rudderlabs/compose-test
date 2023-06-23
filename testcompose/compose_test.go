package testcompose_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/compose-test/compose"
	"github.com/rudderlabs/compose-test/testcompose"
)

const startTimeout = 2 * time.Minute

func TestComposeTesting(t *testing.T) {
	t.Parallel()

	c := testcompose.New(t, compose.FilePaths([]string{"../compose/testdata/docker-compose.test.yml"}))

	ctx, cancel := context.WithTimeout(context.Background(), startTimeout)
	defer cancel()

	c.Start(ctx)

	t.Run("test postgres", func(t *testing.T) {
		t.Parallel()

		port := c.Port("postgresDB", 5432)
		require.NotEqual(t, 5432, port)
		require.NotEqual(t, 0, port)

		user := c.Env("postgresDB", "POSTGRES_USER")
		password := c.Env("postgresDB", "POSTGRES_PASSWORD")

		dbURL := fmt.Sprintf("postgres://%s:%s@localhost:%d/postgres?sslmode=disable",
			user,
			password,
			port,
		)

		conn, err := pgx.Connect(context.Background(), dbURL)
		require.NoError(t, err)
		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "CREATE TABLE test (id int)")
		require.NoError(t, err)

		output := c.Exec(context.Background(), "postgresDB", "psql", "-U", user, "-d", "postgres", "-c", "INSERT INTO test (id) VALUES (1);")
		require.Contains(t, output, "INSERT 0 1")
	})
}
