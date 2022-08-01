package runner

import (
	"io"
	"os"

	"github.com/dpastoor/plr/internal/config"
)

type runOpts struct {
	Id          string
	SessionName string
	Headless    bool
	// default of new session to be true,
	// and if false, need to provide an existing name
	NewSession bool
	// how long before starting the shell session
	Delay  float64
	Stdin  io.ReadCloser
	Stdout io.Writer
	Stderr io.Writer
}

// NewRunOpts sets up the options for a runner with a default
// configuration of creating a new session and wiring up to stdin, stdout, and stderr
func NewDefaultRunOpts(options ...func(*runOpts)) *runOpts {
	opts := &runOpts{NewSession: true}
	opts.Apply(WithInteractiveIO())
	for _, option := range options {
		option(opts)
	}
	return opts
}

// Apply allows a functional option to be applied to a given runOpt instance
func (opts *runOpts) Apply(f func(*runOpts)) {
	f(opts)
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

func WithDelay(delay float64) func(*runOpts) {
	return func(opts *runOpts) {
		opts.Delay = delay
	}
}

// NewOptsFromSession creates a runOpts from a session
func NewOptsFromSession(session config.Session) *runOpts {
	opts := NewDefaultRunOpts()

	if session.Name != nil && *session.Name != "" {
		// if its a explicitly not a new session then we will set byname
		// to also force a non-new session, otherwise use the already applied default
		// of a new session
		if session.New != nil && !*session.New {
			opts.Apply(WithSessionByName(*session.Name))
		} else {
			opts.Apply(WithSessionName(*session.Name))
		}
	}

	if session.Headless != nil && *session.Headless {
		opts.Apply(WithHeadless())
	}
	if session.Id != nil && *session.Id != "" {
		opts.Apply(WithId(*session.Id))
	}
	if session.Delay != nil && *session.Delay > 0 {
		opts.Apply(WithDelay(*session.Delay))
	}
	return opts
}
