package runner

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/metrumresearchgroup/command"
	"github.com/metrumresearchgroup/environ"
)

type runOpts struct {
	Id          string
	SessionName string
	Headless    bool
	// default of new session to be true,
	// and if false, need to provide an existing name
	NewSession bool
	Stdin      io.ReadCloser
	Stdout     io.Writer
	Stderr     io.Writer
}

// NewRunOpts sets up the options for a runner with a default
// configuration of creating a new session and wiring up to stdin, stdout, and stderr
func NewRunOpts(options ...func(*runOpts)) *runOpts {
	opts := &runOpts{NewSession: true}
	WithInteractiveIO()(opts)
	for _, option := range options {
		option(opts)
	}
	return opts
}

// WithNoIO suppresses stdin, stdout, and stderr
func WithNoIO() func(*runOpts) {
	return func(opts *runOpts) {
		opts.Stdin = nil
		opts.Stdout = nil
		opts.Stderr = nil
	}
}

// WithStdin Sets the stdin for the runner
func WithStdin(stdin io.ReadCloser) func(*runOpts) {
	return func(opts *runOpts) {
		opts.Stdin = stdin
	}
}

// WithStdout sets the stdout for the runner
func WithStdout(stdout io.Writer) func(*runOpts) {
	return func(opts *runOpts) {
		opts.Stdout = stdout
	}
}

// WithStderr sets the stderr for the runner
func WithStderr(stderr io.Writer) func(*runOpts) {
	return func(opts *runOpts) {
		opts.Stderr = stderr
	}
}

// WithInteractiveIO returns a runner that will use the stdin, stdout, and stderr
func WithInteractiveIO() func(*runOpts) {
	return func(opts *runOpts) {
		opts.Stdin = os.Stdin
		opts.Stdout = os.Stdout
		opts.Stderr = os.Stderr
	}
}

// WithHeadless sets selenium to run in headless mode
func WithHeadless() func(*runOpts) {
	return func(opts *runOpts) {
		opts.Headless = true
	}
}

// WithSessionName sets the session name and expects a existing session
// therfore sets newsession to false
func WithSessionByName(name string) func(*runOpts) {
	return func(opts *runOpts) {
		opts.NewSession = false
		opts.SessionName = name
	}
}

// WithSessionId sets the session id
func WithId(id string) func(*runOpts) {
	return func(opts *runOpts) {
		opts.Id = id
	}
}

// WithSessionName sets the session name,
// if no session set, will default to the Id
func WithSessionName(sessionName string) func(*runOpts) {
	return func(opts *runOpts) {
		opts.SessionName = sessionName
	}
}

func NewRunner(ctx context.Context, script string, url string, user string, password string, opts *runOpts) *command.Cmd {
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
	return cmd
}
