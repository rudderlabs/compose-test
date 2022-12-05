package compose_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/rudderlabs/compose-test/compose"
	"github.com/stretchr/testify/require"
)

const startTimeout = 2 * time.Minute

func TestCompose(t *testing.T) {
	t.Parallel()

	c, err := compose.New("./testdata/docker-compose.test.yml")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		require.NoError(t, c.Stop(context.Background()))
	})

	ctx, cancel := context.WithTimeout(context.Background(), startTimeout)
	defer cancel()

	require.NoError(t, c.Start(ctx))

	t.Run("test postgres", func(t *testing.T) {
		t.Parallel()

		p, err := c.Port("postgresDB", 5432)
		require.NoError(t, err)
		require.NotEqual(t, 5432, p)
		require.NotEqual(t, 0, p)

		dbName, err := c.Env("postgresDB", "POSTGRES_DB")
		require.NoError(t, err)
		require.Equal(t, "jobsdb", dbName)

		pass, err := c.Env("postgresDB", "POSTGRES_PASSWORD")
		require.NoError(t, err)
		require.Equal(t, "password", pass)

		user, err := c.Env("postgresDB", "POSTGRES_USER")
		require.NoError(t, err)
		require.Equal(t, "rudder", user)

		dbURL := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, pass, p, dbName)

		conn, err := pgx.Connect(context.Background(), dbURL)
		require.NoError(t, err)
		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "CREATE TABLE test (id int)")
		require.NoError(t, err)
	})
}
