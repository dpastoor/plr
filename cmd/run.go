package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/dpastoor/plr/internal/config"
	"github.com/dpastoor/plr/internal/runner"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type runCmd struct {
	cmd  *cobra.Command
	opts runOpts
}

type runOpts struct {
	scenariosPath string
	url           string
	numSessions   int
}

func newRun(runOpts runOpts) error {
	scenarios, err := config.Read(runOpts.scenariosPath)
	url := runOpts.url
	if url == "" {
		url = scenarios.Url
	}
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// defer func() {
	// 	signal.Stop(c)
	// 	cancel()
	// }()
	go func() {
		select {
		case <-c:
			signal.Stop(c)
			fmt.Println("canceling")
			cancel()
		case <-ctx.Done():
			fmt.Println("registering ctx done")
			return
		}
	}()
	wg := &sync.WaitGroup{}
	users := lo.SliceToMap(scenarios.Users, func(user config.User) (string, string) {
		return user.Name, user.Password
	})
	//rand.Shuffle(len(sessions), func(i, j int) { sessions[i], sessions[j] = sessions[j], sessions[i] })
	for i, session := range scenarios.Sessions {
		if i >= runOpts.numSessions {
			continue
		}
		wg.Add(1)
		var delayMs int
		if session.Delay != nil {
			// just in case a negative delay slips
			delayMs = int(math.Max(*session.Delay, 0) * 1000)
		}
		go func(wg *sync.WaitGroup, s config.Session, num int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Duration(delayMs) * time.Millisecond):
				fmt.Printf("launching session %v for user: %s\n", num, s.User)
				opts := runner.NewOptsFromSession(s)
				password, ok := users[s.User]
				if !ok {
					log.Errorf("could not look up password for user %s, not starting session %v", s.User, num)
					return
				}
				runner := runner.NewRunner(ctx, "simple_user_sim.py", url, s.User, password, opts)
				if err := runner.Run(); err != nil {
					log.Errorf("cmd failed to start session %v for user: %s with err %s\n", num, s.User, err)
					return
				}
				log.Infof("completed session %v for user: %s\n", num, s.User)
			}
		}(wg, session, i+1)
	}
	wg.Wait()
	fmt.Println("done waiting on sessions to finish/cleanup")
	return ctx.Err()
}

func setRunOpts(runOpts *runOpts) {
	runOpts.scenariosPath = viper.GetString("scenarios-path")
	runOpts.url = viper.GetString("url")
	numSessions := viper.GetInt("num-sessions")
	if numSessions == 0 {
		runOpts.numSessions = math.MaxInt
	} else {
		runOpts.numSessions = numSessions
	}
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
				log.Fatalf("failed to complete all runs with err: %s", err)
			}
			return nil
		},
	}
	cmd.Flags().String("scenarios-path", "scenarios.json", "path to scenarios file")
	viper.BindPFlag("scenarios-path", cmd.Flags().Lookup("scenarios-path"))
	cmd.Flags().IntP("num-sessions", "n", 0, "number of sessions to run")
	viper.BindPFlag("num-sessions", cmd.Flags().Lookup("num-sessions"))
	cmd.Flags().String("url", "", "path to server")
	viper.BindPFlag("url", cmd.Flags().Lookup("url"))
	root.cmd = cmd

	return root
}
