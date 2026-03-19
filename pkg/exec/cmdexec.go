package exec

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

type CmdExecutor struct {
	log *slog.Logger
	c   func(ctx context.Context, name string, arg ...string) *exec.Cmd
}

type Params struct {
	Name     string
	Args     []string
	Dir      string
	Timeout  time.Duration
	Combined bool
	Stdin    string
	Env      []string
}

func New(log *slog.Logger) *CmdExecutor {
	return &CmdExecutor{
		log: log,
		c:   exec.CommandContext,
	}
}

func (i *CmdExecutor) WithCommandFn(c func(ctx context.Context, name string, arg ...string) *exec.Cmd) *CmdExecutor {
	i.c = c
	return i
}

func (i *CmdExecutor) Execute(ctx context.Context, p *Params) (out string, err error) {
	var (
		start  = time.Now()
		output []byte
	)
	i.log.Debug("running command", "command", strings.Join(append([]string{p.Name}, p.Args...), " "), "start", start.String())

	if p.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}

	cmd := i.c(ctx, p.Name, p.Args...)
	cmd.Env = append(cmd.Env, p.Env...)

	// show stderr
	cmd.Stderr = os.Stderr

	if p.Stdin != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", err
		}

		go func() {
			defer func() {
				_ = stdin.Close()
			}()
			_, err = io.WriteString(stdin, p.Stdin)
			if err != nil {
				i.log.Error("error when writing to command's stdin", "error", err)
			}
		}()
	}

	if p.Combined {
		output, err = cmd.CombinedOutput()
	} else {
		output, err = cmd.Output()
	}

	out = string(output)
	took := time.Since(start)

	if err != nil {
		i.log.Error("executed command with error", "output", out, "duration", took.String(), "error", err)
		return "", err
	}

	i.log.Info("executed command", "output", out, "duration", took.String())

	return
}
