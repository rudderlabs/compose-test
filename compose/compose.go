package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type TestingCompose struct {
	compose *Compose
	t       testing.TB
}

type Compose struct {
	ports map[string]map[int]int
	name  string
	paths []string
}

type containerInfo struct {
	Service    string
	Publishers []publisher
}

// Publishers":[{"URL":"","TargetPort":8123,"PublishedPort":0,"Protocol":"tcp"}
type publisher struct {
	Protocol      string
	URL           string
	TargetPort    int
	PublishedPort int
}

func Open(paths ...string) (*Compose, error) {
	return &Compose{
		name:  "test_" + randSeq(6),
		paths: paths,
	}, nil
}

func OpenTesting(t testing.TB, paths ...string) *TestingCompose {
	c, err := Open(paths...)
	if err != nil {
		t.Fatalf("open compose: %v", err)
	}

	return &TestingCompose{
		compose: c,
		t:       t,
	}
}

func (c *Compose) Stop(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker",
		"compose",
		"-p", c.name,

		"down",
		"--timeout", "0",
		"--rmi", "local",
		"--volumes",
	)
	o, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(o))
		return fmt.Errorf("docker compose down: %w", err)
	}

	return nil
}

func (tc *TestingCompose) Stop(ctx context.Context) {
	err := tc.compose.Stop(ctx)
	if err != nil {
		tc.t.Fatalf("compose library stop: %v", err)
	}
}

func (c *Compose) Start(ctx context.Context) error {
	args := []string{
		"compose",
		"-p", c.name,
	}

	for _, path := range c.paths {
		args = append(args, "-f", path)
	}

	args = append(args, "up", "--detach", "--wait")

	cmd := exec.CommandContext(ctx, "docker", args...)
	o, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(args)
		fmt.Println(string(o))
		return fmt.Errorf("docker compose up: %w", err)
	}

	err = c.extractPorts(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (tc *TestingCompose) Start(ctx context.Context) {
	err := tc.compose.Start(ctx)
	if err != nil {
		tc.t.Fatalf("compose library start: %v", err)
	}
}

func (c *Compose) Port(service string, port int) (int, error) {
	s, ok := c.ports[service]
	if !ok {
		return 0, fmt.Errorf("no service %q found", service)
	}

	p, ok := s[port]
	if !ok {
		return 0, fmt.Errorf("port %d is not published", port)
	}

	return p, nil
}

func (tc *TestingCompose) Port(service string, port int) int {
	p, err := tc.compose.Port(service, port)
	if err != nil {
		tc.t.Fatalf("compose library port: %v", err)
	}

	return p
}

func (c *Compose) extractPorts(ctx context.Context) error {
	info, err := c.ps(ctx)
	if err != nil {
		return err
	}

	c.ports = make(map[string]map[int]int)

	for _, i := range info {
		p := make(map[int]int)
		for _, pub := range i.Publishers {
			p[pub.TargetPort] = pub.PublishedPort
		}
		c.ports[i.Service] = p
	}

	return nil
}

func (c *Compose) ps(ctx context.Context) ([]containerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker",
		"compose",
		"-p", c.name,

		"ps",
		"--format", "json",
	)
	o, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(o))
		return nil, fmt.Errorf("docker compose ps: %w", err)
	}

	var info []containerInfo
	err = json.Unmarshal(o, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
