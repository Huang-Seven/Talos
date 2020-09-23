package main

import (
	"Talos/internal/models"
	"log"
	"os"

	"github.com/urfave/cli"
)

const (
	DefaultSrvHost = "127.0.0.1"
	DefaultSrvPort = 50050
)

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = "Talos"
	cliApp.Usage = "Talos client"
	cliApp.Version = "0.1"
	cliApp.Commands = getCommands()
	cliApp.Flags = append(cliApp.Flags, []cli.Flag{}...)
	err := cliApp.Run(os.Args)
	if err != nil {
		//logger.Fatal(err)
	}
}

// getCommands
func getCommands() []cli.Command {
	command := cli.Command{
		Name:   "run",
		Usage:  "run talos client",
		Action: run,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "srvhost",
				Value: DefaultSrvHost,
				Usage: "bind server host",
			},
			cli.IntFlag{
				Name:  "srvport",
				Value: DefaultSrvPort,
				Usage: "bind port",
			},
			cli.StringFlag{
				Name:  "cd",
				Value: "/etc/monitor_shell/",
				Usage: "bind conf dir",
			},
		},
	}

	return []cli.Command{command}
}

func run(ctx *cli.Context) {
	a := new(models.Agent)
	a.Ac.ServerAddr = ctx.String("srvhost")
	a.Ac.ServerPort = ctx.Int("srvport")
	if ctx.IsSet("cd") {
		a.Ac.ConfDir = ctx.String("cd")
	}
	log.Printf("Agent version: %v", ctx.App.Version)
	a.Run()
}
