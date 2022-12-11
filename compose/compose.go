package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
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

type Compose struct {
	file   File
	ports  map[string]map[int]int
	env    map[string]map[string]string
	labels map[string]string

	name string
}

type containerInfo struct {
	ID         string
	Service    string
	Publishers []publisher
}

type containerConfig struct {
	Env    []string          `json:"Env"`
	Labels map[string]string `json:"Labels"`
}

// Publishers":[{"URL":"","TargetPort":8123,"PublishedPort":0,"Protocol":"tcp"}
type publisher struct {
	Protocol      string
	URL           string
	TargetPort    int
	PublishedPort int
}

func New(file File) (*Compose, error) {
	return &Compose{
		name: "test_" + randSeq(20),
		file: file,
	}, nil
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
		fmt.Fprintln(os.Stderr, string(o))
		return fmt.Errorf("docker compose down: %w", err)
	}

	return nil
}

func (c *Compose) Start(ctx context.Context) error {
	args := []string{
		"compose",
		"-p", c.name,
	}

	cmd := exec.CommandContext(ctx, "docker", args...)

	c.file.apply(cmd)

	cmd.Args = append(cmd.Args, "up", "--detach", "--wait")

	o, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, cmd.String())
		fmt.Fprintln(os.Stderr, string(o))
		return fmt.Errorf("docker compose up: %w: %s", err, string(o))
	}

	err = c.extractServiceInfo(ctx)
	if err != nil {
		return err
	}

	return nil
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

func (c *Compose) Env(service, name string) (string, error) {
	kv, ok := c.env[service]
	if !ok {
		return "", fmt.Errorf("no service %q found", service)
	}

	v, ok := kv[name]
	if !ok {
		return "", fmt.Errorf("environment variable %q is not published", name)
	}

	return v, nil
}

func (c *Compose) extractServiceInfo(ctx context.Context) error {
	psInfo, err := c.ps(ctx)
	if err != nil {
		return err
	}

	c.ports = make(map[string]map[int]int)

	ids := make([]string, len(psInfo))
	for i, info := range psInfo {
		p := make(map[int]int)
		for _, pub := range info.Publishers {
			p[pub.TargetPort] = pub.PublishedPort
		}
		c.ports[info.Service] = p
		ids[i] = info.ID
	}

	configs, err := c.config(ctx, ids...)
	if err != nil {
		return err
	}

	c.env = make(map[string]map[string]string)

	for i, config := range configs {
		c.env[psInfo[i].Service] = make(map[string]string)
		for _, env := range config.Env {
			s := strings.SplitN(env, "=", 2)
			c.env[psInfo[i].Service][s[0]] = s[1]
		}

		for _, label := range config.Labels {
			if strings.HasPrefix(label, "compose-test-expose") {
				s := strings.SplitN(
					strings.TrimPrefix(label, "compose-test-expose"),
					"=",
					2,
				)
				c.labels[s[0]] = s[1]
			}
		}

	}

	return nil
}

func (c *Compose) config(ctx context.Context, id ...string) ([]containerConfig, error) {
	args := []string{
		"inspect",
		"--format={{json .Config}}",
	}
	args = append(args, id...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Env = os.Environ() // TODO: remove this
	o, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker compose ps: %w: %s", err, o)
	}

	lines := strings.Split(strings.TrimSpace(string(o)), "\n")

	configs := make([]containerConfig, len(lines))
	for i, l := range lines {
		err = json.Unmarshal([]byte(l), &configs[i])
		if err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
	}

	return configs, nil
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
		return nil, fmt.Errorf("docker compose ps: %w", err)
	}

	var info []containerInfo
	err = json.Unmarshal(o, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
