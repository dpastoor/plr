package cmd

import (
	"fmt"

	"github.com/dpastoor/plr/internal/runner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runCmd struct {
	cmd  *cobra.Command
	opts runOpts
}

type runOpts struct {
}

func newRun(runOpts runOpts) error {
	opts := runner.NewRunOpts()
	return runner.NewRunner("simple_user_sim.py", "http://localhost:8787", "dpastoor", "password123", opts)
}

func setRunOpts(runOpts *runOpts) {

}

func (opts *runOpts) Validate() error {
	return nil
}

func newRunCmd() *runCmd {
	root := &runCmd{opts: runOpts{}}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "run",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			setRunOpts(&root.opts)
			if err := root.opts.Validate(); err != nil {
				return err
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			//TODO: Add your logic to gather config to pass code here
			log.WithField("opts", fmt.Sprintf("%+v", root.opts)).Trace("run-opts")
			if err := newRun(root.opts); err != nil {
				return err
			}
			return nil
		},
	}
	root.cmd = cmd
	return root
}
