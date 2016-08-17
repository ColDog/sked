package main

import (
	"github.com/coldog/scheduler/actions"
	. "github.com/coldog/scheduler/agent"
	. "github.com/coldog/scheduler/api"
	. "github.com/coldog/scheduler/scheduler"

	"github.com/urfave/cli"
	log "github.com/Sirupsen/logrus"

	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func NewApp() *App {
	app := &App{
		cli:    cli.NewApp(),
		Config: &Config{},
	}
	app.setup()
	return app
}

type ExitHandler func()
type AppCmd func(app *App) cli.Command

type App struct {
	cli    *cli.App
	Api    *SchedulerApi
	Agent  *Agent
	Master *Master
	Config *Config
	atExit ExitHandler
}

func (app *App) printWelcome(mode string) {
	fmt.Printf("\nconsul-scheduler starting...\n\n")
	fmt.Printf("          version  %s\n", VERSION)
	fmt.Printf("        log-level  %s\n", log.GetLevel())
	fmt.Printf("       consul-api  %s\n", app.Api.ConsulConf.Address)
	fmt.Printf("        consul-dc  %s\n", app.Api.ConsulConf.Datacenter)
	fmt.Printf("             mode  %s\n", mode)
	fmt.Printf("             pid   %d\n", os.Getpid())
	fmt.Print("\nlog output will now begin streaming....\n")
}

func (app *App) setup() {
	app.cli.Name = "consul-scheduler"
	app.cli.Version = VERSION
	app.cli.Author = "Colin Walker"
	app.cli.Usage = "schedule tasks across a consul cluster."

	app.cli.Flags = []cli.Flag{
		cli.StringFlag{Name: "log, l", Value: "debug", Usage: "log level [debug, info, warn, error]"},
		cli.StringFlag{Name: "consul-api", Value: "", Usage: "consul api"},
		cli.StringFlag{Name: "consul-dc", Value: "", Usage: "consul dc"},
		cli.StringFlag{Name: "consul-token", Value: "", Usage: "consul token"},
	}
	app.cli.Before = func(c *cli.Context) error {
		app.Config.ConsulApiAddress = c.GlobalString("consul-api")
		app.Config.ConsulApiDc = c.GlobalString("consul-dc")
		app.Config.ConsulApiToken = c.GlobalString("consul-token")
		app.Api = NewSchedulerApiWithConfig(app.Config)

		switch c.GlobalString("log-level") {
		case "debug":
			log.SetLevel(log.DebugLevel)
			break
		case "info":
			log.SetLevel(log.InfoLevel)
			break
		case "warn":
			log.SetLevel(log.WarnLevel)
			break
		case "error":
			log.SetLevel(log.ErrorLevel)
			break
		default:
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}

	app.cli.Commands = []cli.Command{
		app.AgentCmd(),
		app.SchedulerCmd(),
		app.CombinedCmd(),
		app.ApplyCmd(),
	}
}


// Registers a listener for SIGTERM and sets the function to run being the 'ExitHandler' function.
func (app *App) AtExit(e ExitHandler) {
	app.atExit = e

	killCh := make(chan os.Signal, 2)
	signal.Notify(killCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-killCh

		fmt.Println(" caught interrupt ")

		// ensure that we will actually quit within 10 seconds, but allow for some
		// cleanup to happen by the code before we exit.
		go func() {
			time.Sleep(10 * time.Second)
			log.Fatal("[main] failed to exit cleanly")
		}()

		app.atExit()
	}()
}

func (app *App) RegisterAgent() {
	app.Agent = NewAgent(app.Api)
}

func (app *App) RegisterMaster() {
	app.Master = NewMaster(app.Api)
}

func (app *App) AgentCmd() (cmd cli.Command) {
	cmd.Name = "agent"
	cmd.Usage = "start the agent service"
	cmd.Action = func(c *cli.Context) error {
		app.printWelcome("agent")
		app.Api.WaitForInitialize()

		app.RegisterAgent()
		app.AtExit(func() {
			app.Agent.Stop()
		})

		app.Agent.Run()
		return nil
	}
	return cmd
}

func (app *App) SchedulerCmd() (cmd cli.Command) {
	cmd.Name = "scheduler"
	cmd.Usage = "start the scheduler service"
	cmd.Action = func(c *cli.Context) error {
		app.printWelcome("scheduler")
		app.Api.WaitForInitialize()

		app.RegisterMaster()
		app.AtExit(func() {
			app.Master.Stop()
		})

		app.Master.Run()
		return nil
	}
	return cmd
}

func (app *App) CombinedCmd() (cmd cli.Command) {
	cmd.Name = "combined"
	cmd.Usage = "start the scheduler and agent service"
	cmd.Action = func(c *cli.Context) error {
		app.printWelcome("combined")
		app.Api.WaitForInitialize()

		app.RegisterMaster()
		app.RegisterAgent()

		wg := &sync.WaitGroup{}

		wg.Add(2)
		go func() {
			app.Agent.Run()
			wg.Done()
		}()

		go func() {
			app.Master.Run()
			wg.Done()
		}()

		app.AtExit(func() {
			app.Agent.Stop()
			app.Master.Stop()
		})

		wg.Wait()
		return nil
	}
	return cmd
}

func (app *App) ApplyCmd() (cmd cli.Command) {
	cmd.Name = "apply"
	cmd.Usage = "apply the provided configuration file to consul"
	cmd.Flags = []cli.Flag{
		cli.StringFlag{Name: "file, f", Value: "consul-scheduler.yml", Usage: "configuration file to read"},
	}
	cmd.Action = func(c *cli.Context) error {
		file := c.String("file")
		return actions.ApplyConfig(file, app.Api)
	}
	return cmd
}

func (app *App) AddCmd(cmd AppCmd) {
	app.cli.Commands = append(app.cli.Commands, cmd(app))
}

func (app *App) Run() {
	app.cli.Run(os.Args)
}

func main() {
	app := NewApp()
	app.Run()
}
