package runner

import (
	"context"
	"fmt"

	"github.com/metrumresearchgroup/command"
	"github.com/metrumresearchgroup/environ"
)

type Runner struct {
	cmd *command.Cmd
}

func NewRunner(ctx context.Context, script string, url string, user string, password string, opts *runOpts) *Runner {
	env := environ.FromOS()
	cmdArgs := []string{
		script,
		fmt.Sprintf("--url=%s", url),
		fmt.Sprintf("--user=%s", user),
		fmt.Sprintf("--password=%s", password),
	}
	if opts.Headless {
		cmdArgs = append(cmdArgs, "--headless")
	}
	if opts.NewSession {
		cmdArgs = append(cmdArgs, "--new-session")
	}
	if opts.Id != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--id=%s", opts.Id))
	}
	if opts.SessionName != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--session-name=%s", opts.SessionName))
	}

	cmd := command.NewWithContext(ctx, "python", cmdArgs...)
	cmd.Env = env.AsSlice()
	command.WireIO(opts.Stdin, opts.Stdout, opts.Stderr).Apply(cmd)
	return &Runner{
		cmd: cmd,
	}
}

func (r *Runner) Run() error {
	return r.cmd.Run()
}
