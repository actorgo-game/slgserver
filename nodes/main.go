package main

import (
	"fmt"
	"os"

	cherryConst "github.com/actorgo-game/actorgo/const"
	"github.com/llr104/slgserver/internal/node/center"
	"github.com/llr104/slgserver/internal/node/chat"
	"github.com/llr104/slgserver/internal/node/game"
	"github.com/llr104/slgserver/internal/node/gate"
	"github.com/llr104/slgserver/internal/node/master"
	"github.com/llr104/slgserver/internal/node/web"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "slgserver",
		Description: "SLG game server powered by actorgo",
		Commands: []*cli.Command{
			versionCommand(),
			masterCommand(),
			centerCommand(),
			webCommand(),
			gateCommand(),
			gameCommand(),
			chatCommand(),
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
			fmt.Println(cherryConst.Version())
			return nil
		},
	}
}

func masterCommand() *cli.Command {
	return &cli.Command{
		Name:  "master",
		Usage: "run master node (service discovery)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			master.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func centerCommand() *cli.Command {
	return &cli.Command{
		Name:  "center",
		Usage: "run center node (account management)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			center.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func webCommand() *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "run web node (HTTP API)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			web.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func gateCommand() *cli.Command {
	return &cli.Command{
		Name:  "gate",
		Usage: "run gate node (WebSocket gateway)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			gate.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func gameCommand() *cli.Command {
	return &cli.Command{
		Name:  "game",
		Usage: "run game node (game logic)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			game.Run(c.String("path"), c.String("node"))
			return nil
		},
	}
}

func chatCommand() *cli.Command {
	return &cli.Command{
		Name:  "chat",
		Usage: "run chat node",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			chat.Run(c.String("path"), c.String("node"))
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
