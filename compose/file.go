package compose

import (
	"bytes"
	"os/exec"
)

type File interface {
	apply(c *exec.Cmd)
}

type FilePath string

func (f FilePath) apply(c *exec.Cmd) {
	c.Args = append(c.Args, "-f", string(f))
}

type FileBytes []byte

func (f FileBytes) apply(c *exec.Cmd) {
	c.Args = append(c.Args, "-f", "/dev/stdin")
	c.Stdin = bytes.NewReader(f)
}
