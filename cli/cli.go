package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
)

type ProcInout struct {
	Stdin  io.ReadCloser
	Stdout io.WriteCloser
	Stderr io.WriteCloser
	Env    EnvFunc
}

type EnvFunc func(name string) string

func NewProcInout() *ProcInout {
	return &ProcInout{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    os.Getenv,
	}
}

func StubProcInout() *ProcInout {
	return &ProcInout{
		Stdin:  io.NopCloser(strings.NewReader("")),
		Stdout: NewNopWriteCloser(io.Discard),
		Stderr: NewNopWriteCloser(io.Discard),
		Env:    func(name string) string { return "" },
	}
}

type ProcInoutSpy struct {
	Stdin  io.Reader
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
	Env    map[string]string
}

func (s *ProcInoutSpy) NewProcInout() *ProcInout {
	return &ProcInout{
		Stdin:  io.NopCloser(s.Stdin),
		Stdout: NewNopWriteCloser(s.Stdout),
		Stderr: NewNopWriteCloser(s.Stderr),
		Env:    NewEnvFunc(s.Env),
	}
}

func SpyProcInout(stdin ...string) *ProcInoutSpy {
	return &ProcInoutSpy{
		Stdin:  strings.NewReader(strings.Join(stdin, "\n")),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Env:    make(map[string]string),
	}
}

func OpenInput(name string, inout *ProcInout) (io.ReadCloser, error) {
	if name == "-" {
		return inout.Stdin, nil
	}
	return os.Open(name)
}

type Command func(args []string, inout *ProcInout) int

func Run(c Command) {
	args := os.Args[1:]
	exitStatus := c(args, NewProcInout())
	os.Exit(exitStatus)
}

func NewEnvFunc(env map[string]string) EnvFunc {
	return func(name string) string {
		return env[name]
	}
}
