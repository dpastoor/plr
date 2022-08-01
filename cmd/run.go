package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

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
	numLaunched := 0
	var users []string
	for i := 11; i < 35; i++ {
		users = append(users, fmt.Sprintf("user%d", i))
	}
	//var sessions []string
	wg := &sync.WaitGroup{}
	//rand.Shuffle(len(sessions), func(i, j int) { sessions[i], sessions[j] = sessions[j], sessions[i] })
loop:
	for i := 0; i < 6; i++ {
		for _, user := range users {
			select {
			case <-ctx.Done():
				fmt.Println("got to ctx done")
				break loop
			default:
				numLaunched += 1
				wg.Add(1)
				go func(wg *sync.WaitGroup, sessionUser string, num int) {
					defer wg.Done()
					select {
					case <-ctx.Done():
						return
					case <-time.Tick(time.Duration(rand.Intn(100000)) * time.Millisecond):
						fmt.Printf("launching session %v for user: %s\n", num, sessionUser)
						opts := runner.NewDefaultRunOpts(runner.WithHeadless())
						runner := runner.NewRunner(ctx, "simple_user_sim.py", "http://ec2-18-117-188-179.us-east-2.compute.amazonaws.com:8787", sessionUser, "password123", opts)
						if err := runner.Run(); err != nil {
							fmt.Printf("cmd failed to start session %v for user: %s with err %s\n", num, sessionUser, err)
							return
						}
						fmt.Printf("completed session %v for user: %s\n", num, sessionUser)
					}
				}(wg, user, numLaunched)
			}
		}
		select {
		case <-time.Tick(time.Second * 150):
			continue
		case <-ctx.Done():
			break loop
		}
	}
	wg.Wait()
	fmt.Println("done waiting on sessions to finish/cleanup")
	return ctx.Err()
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
				log.Fatalf("failed to complete all runs with err: %s", err)
			}
			return nil
		},
	}
	root.cmd = cmd
	return root
}
