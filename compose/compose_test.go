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

	c, err := compose.Open("./testdata/docker-compose.test.yml")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), startTimeout)
	defer cancel()

	require.NoError(t, c.Start(ctx))

	t.Run("test postgres", func(t *testing.T) {
		t.Parallel()

		p, err := c.Port("wh-jobsDb", 5432)
		require.NoError(t, err)
		require.NotEqual(t, 5432, p)
		require.NotEqual(t, 0, p)

		dbURL := fmt.Sprintf("postgres://rudder:password@localhost:%d/postgres?sslmode=disable", p)

		conn, err := pgx.Connect(context.Background(), dbURL)
		require.NoError(t, err)
		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "CREATE TABLE test (id int)")
		require.NoError(t, err)
	})
}

func TestComposeTesting(t *testing.T) {
	t.Parallel()

	c := compose.OpenTesting(t, "./testdata/docker-compose.test.yml")

	ctx, cancel := context.WithTimeout(context.Background(), startTimeout)
	defer cancel()

	c.Start(ctx)

	t.Run("test postgres", func(t *testing.T) {
		t.Parallel()

		p := c.Port("wh-jobsDb", 5432)
		require.NotEqual(t, 5432, p)
		require.NotEqual(t, 0, p)

		dbURL := fmt.Sprintf("postgres://rudder:password@localhost:%d/postgres?sslmode=disable", p)

		conn, err := pgx.Connect(context.Background(), dbURL)
		require.NoError(t, err)
		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "CREATE TABLE test (id int)")
		require.NoError(t, err)
	})
}
