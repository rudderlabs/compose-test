package testcompose

import (
	"context"
	"testing"

	"github.com/rudderlabs/compose-test/compose"
)

type TestingCompose struct {
	compose *compose.Compose
	t       testing.TB
}

func New(t testing.TB, paths ...string) *TestingCompose {
	c, err := compose.New(paths...)
	if err != nil {
		t.Fatalf("open compose: %v", err)
	}

	return &TestingCompose{
		compose: c,
		t:       t,
	}
}

func (tc *TestingCompose) Start(ctx context.Context) {
	err := tc.compose.Start(ctx)
	if err != nil {
		tc.t.Fatalf("compose library start: %v", err)
	}

	tc.t.Cleanup(func() {
		tc.Stop(context.Background())
	})
}

func (tc *TestingCompose) Stop(ctx context.Context) {
	err := tc.compose.Stop(ctx)
	if err != nil {
		tc.t.Fatalf("compose library stop: %v", err)
	}
}

func (tc *TestingCompose) Port(service string, port int) int {
	p, err := tc.compose.Port(service, port)
	if err != nil {
		tc.t.Fatalf("compose library port: %v", err)
	}

	return p
}

func (tc *TestingCompose) Env(service string, name string) string {
	v, err := tc.compose.Env(service, name)
	if err != nil {
		tc.t.Fatalf("compose library port: %v", err)
	}

	return v
}
