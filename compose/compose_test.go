package compose_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rudderlabs/compose-test/compose"
	"github.com/stretchr/testify/require"
)

func TestCompose(t *testing.T) {
	t.Parallel()

	c, err := compose.Open("./testdata/docker-compose.test.yml")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Cleanup(func() {
		if err := c.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Cleanup(func() {
		c.Stop(context.Background())
	})

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
