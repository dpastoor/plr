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
	scriptPath    string
	url           string
	numSessions   int
	unique        bool
	noDelay       bool
	python        string
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
	hasRunForUser := make(map[string]bool)
	for user := range users {
		hasRunForUser[user] = false
	}
	//rand.Shuffle(len(sessions), func(i, j int) { sessions[i], sessions[j] = sessions[j], sessions[i] })
	for i, session := range scenarios.Sessions {
		if runOpts.noDelay {
			session.Delay = nil
		}
		if i >= runOpts.numSessions {
			continue
		}
		if runOpts.unique {
			if hasRunForUser[session.User] {
				continue
			}
			hasRunForUser[session.User] = true
		}
		wg.Add(1)
		var delayMs int
		if session.Delay == nil {
			delayMs = 5
		} else {
			delayMs = int(math.Max(*session.Delay, 0)*1000) + 5
		}
		go func(wg *sync.WaitGroup, s config.Session, num int) {
			startTime := time.Now()
			log.Infof("queued session %v for user: %s after %.3f seconds since start\n", num, s.User, time.Since(startTime).Seconds())
			defer wg.Done()
			select {
			case <-ctx.Done():
				log.Warnf("context done for session %d before starting", num)
				return
			case <-time.Tick(time.Duration(delayMs) * time.Millisecond):
				log.Printf("launching session %v for user: %s after %.3f seconds since start\n", num, s.User, time.Since(startTime).Seconds())
				opts := runner.NewOptsFromSession(s)
				opts.Apply(runner.WithPythonPath(runOpts.python))
				password, ok := users[s.User]
				if !ok {
					log.Errorf("could not look up password for user %s, not starting session %v", s.User, num)
					return
				}
				runner := runner.NewRunner(ctx, runOpts.scriptPath, url, s.User, password, s.RemoteCmdBase64, opts)
				if err := runner.Run(); err != nil {
					log.Errorf("cmd failed to start session %v for user: %s with err %s\n", num, s.User, err)
					return
				}
				log.Infof("completed session %v for user: %s\n", num, s.User)
			}
		}(wg, session, i+1)
	}
	wg.Wait()
	log.Info("done waiting on sessions to finish/cleanup")
	return ctx.Err()
}

func setRunOpts(runOpts *runOpts, args []string) {
	runOpts.scenariosPath = viper.GetString("scenarios-path")
	runOpts.url = viper.GetString("url")
	numSessions := viper.GetInt("num-sessions")
	if numSessions == 0 {
		runOpts.numSessions = math.MaxInt
	} else {
		runOpts.numSessions = numSessions
	}
	if len(args) != 1 {
		log.Fatal("must specify script to run")
	}
	runOpts.scriptPath = args[0]
	runOpts.unique = viper.GetBool("unique")
	runOpts.noDelay = viper.GetBool("no-delay")
	runOpts.python = viper.GetString("python")
}

func (opts *runOpts) Validate() error {
	_, err := os.Open(opts.scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("script file %s does not exist", opts.scriptPath)
		} else {
			return fmt.Errorf("could not open script file %s with err %s", opts.scriptPath, err)
		}

	}
	return nil
}

func newRunCmd() *runCmd {
	root := &runCmd{opts: runOpts{}}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "run <path/to/python/script>",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			setRunOpts(&root.opts, args)
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
	cmd.Flags().Bool("unique", false, "run sessions for each user only once")
	viper.BindPFlag("unique", cmd.Flags().Lookup("unique"))
	cmd.Flags().Bool("no-delay", false, "start immediately instead of waiting for delay")
	viper.BindPFlag("no-delay", cmd.Flags().Lookup("no-delay"))

	cmd.Flags().String("python", "python", "path to python executable")
	viper.BindPFlag("python", cmd.Flags().Lookup("python"))

	viper.SetEnvPrefix("PLR")
	viper.BindEnv("python")
	root.cmd = cmd

	return root
}
