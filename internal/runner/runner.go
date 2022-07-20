package runner

import (
	"fmt"

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
}

func NewRunOpts(options ...func(*runOpts)) *runOpts {
	opts := &runOpts{NewSession: true}
	for _, option := range options {
		option(opts)
	}
	return opts
}

func WithHeadless(opts *runOpts) func(*runOpts) {
	return func(opts *runOpts) {
		opts.Headless = true
	}
}

// WithSessionName sets the session name and expects a existing session
// therfore sets newsession to false
func WithSessionByName(opts *runOpts, name string) func(*runOpts) {
	return func(opts *runOpts) {
		opts.NewSession = false
		opts.SessionName = name
	}
}

func WithId(opts *runOpts, id string) func(*runOpts) {
	return func(opts *runOpts) {
		opts.Id = id
	}
}

func WithSessionName(opts *runOpts, sessionName string) func(*runOpts) {
	return func(opts *runOpts) {
		opts.SessionName = sessionName
	}
}

func NewRunner(script string, url string, user string, password string, opts *runOpts) error {
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

	cmd := command.New("python", cmdArgs...)
	cmd.Env = env.AsSlice()
	command.InteractiveIO().Apply(cmd)
	// using our command package
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
