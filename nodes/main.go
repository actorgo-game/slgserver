package main

import (
	"fmt"
	"os"

	cconst "github.com/actorgo-game/actorgo/const"
	"github.com/llr104/slgserver/internal/node/chatserver"
	"github.com/llr104/slgserver/internal/node/gateserver"
	"github.com/llr104/slgserver/internal/node/httpserver"
	"github.com/llr104/slgserver/internal/node/loginserver"
	"github.com/llr104/slgserver/internal/node/slgserver"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "slgserver",
		Description: "SLG slgserver server powered by actorgo",
		Commands: []*cli.Command{
			versionCommand(),
			loginserverCommand(),
			httpserverCommand(),
			gateserverCommand(),
			slgserverCommand(),
			chatserverCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"ver", "v"},
		Usage:   "view version",
		Action: func(c *cli.Context) error {
			fmt.Println(cconst.Version())
			return nil
		},
	}
}

func loginserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "loginserver",
		Usage: "run loginserver node (account management)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			loginserver.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func httpserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "httpserver",
		Usage: "run httpserver node (HTTP API)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			httpserver.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func gateserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "gateserver",
		Usage: "run gateserver node (WebSocket gateway)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			gateserver.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func slgserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "slgserver",
		Usage: "run slgserver node (slgserver logic)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			slgserver.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func chatserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "chatserver",
		Usage: "run chatserver node",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			chatserver.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "path",
			Usage:    "profile config file path",
			Required: false,
			Value:    "./config/profile-dev.json",
		},
		&cli.StringFlag{
			Name:     "node",
			Usage:    "node id",
			Required: true,
		},
	}
}
