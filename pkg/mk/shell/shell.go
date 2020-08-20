package shell

import (
	"bufio"
	"context"
	"github.com/rs/zerolog"
	"io"
	"os/exec"
)

var DefaultShell = []string{"/usr/bin/env", "sh", "-c"}

type ShellExecutor struct {
	ShellCmd []string
	Dir string
	Env []string
}

func (s *ShellExecutor) shell() []string {
	if s.ShellCmd == nil {
		return DefaultShell
	}
	return s.ShellCmd
}

func (s *ShellExecutor) RunShell(ctx context.Context, cmd string) error {
	log := zerolog.Ctx(ctx)
	shell := s.shell()
	args := append([]string{}, shell[1:]...)
	args = append(args, cmd)
	c := exec.Command(shell[0], args...)
	if s.Dir != "" {
		c.Dir = s.Dir
	}
	if len(s.Env) > 0 {
		c.Env = s.Env
	}
	lw := logWriter(log)
	c.Stderr = lw
	c.Stdout = lw
	return c.Run()
}

func logWriter(l *zerolog.Logger) io.WriteCloser {
	pr, pw := io.Pipe()
	scanner := bufio.NewScanner(pr)
	go func() {
		for scanner.Scan() {
			l.Trace().Msg(scanner.Text())
		}
	}()
	return pw
}