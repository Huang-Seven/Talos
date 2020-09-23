package main

import (
	"Talos/internal/models"
	"log"
	"os"

	"github.com/urfave/cli"
)

const (
	DefaultPort = 50050
)

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = "Talos"
	cliApp.Usage = "Talos server"
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
		Usage:  "run talos server",
		Action: run,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "host",
				Value: "0.0.0.0",
				Usage: "bind host",
			},
			cli.IntFlag{
				Name:  "port",
				Value: DefaultPort,
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
	s := new(models.Server)
	s.Sc.LocalPort = ctx.Int("port")
	if ctx.IsSet("cd") {
		s.Sc.ConfDir = ctx.String("cd")
	}
	log.Printf("Server version: %v", ctx.App.Version)
	s.Run()
}
