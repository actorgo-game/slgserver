package main

import (
	"fmt"
	"os"

	cconst "github.com/actorgo-game/actorgo/const"
	cstring "github.com/actorgo-game/actorgo/extend/string"
	cfacade "github.com/actorgo-game/actorgo/facade"
	"github.com/llr104/slgserver/internal/node/chatserver"
	"github.com/llr104/slgserver/internal/node/gateserver"
	"github.com/llr104/slgserver/internal/node/httpserver"
	"github.com/llr104/slgserver/internal/node/loginserver"
	"github.com/llr104/slgserver/internal/node/master"
	"github.com/llr104/slgserver/internal/node/slgserver"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "slgserver",
		Description: "SLG slgserver server powered by actorgo",
		Commands: []*cli.Command{
			versionCommand(),
			masterCommand(),
			loginserverCommand(),
			httpserverCommand(),
			gateserverCommand(),
			slgserverCommand(),
			chatserverCommand(),
		},
	}

	var strid = os.Args[1]
	fmt.Printf("Parse Args[%v] strid:%v\n", os.Args[1], strid)

	svrid, err := cfacade.GenNodeIdByStr(strid)
	if err != nil {
		fmt.Printf("ParseSvrdID err:%v\n", err)
		return
	}

	nodeType := cfacade.GetNodeType(svrid)
	fmt.Printf("nodeType:%v\n", nodeType)

	var args []string
	args = append(args, os.Args[0])
	args = append(args, cstring.ToString(nodeType))

	if err := app.Run(args); err != nil {
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

func masterCommand() *cli.Command {
	return &cli.Command{
		Name:  "1",
		Usage: "run 1 node",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			path, _ := getParameters(c)
			master.Run(path, os.Args[1])
			return nil
		},
	}
}

func loginserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "2",
		Usage: "run 2 node (account management)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			path, _ := getParameters(c)
			loginserver.Run(path, os.Args[1])
			return nil
		},
	}
}

func httpserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "3",
		Usage: "run 3 node (HTTP API)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			path, _ := getParameters(c)
			httpserver.Run(path, os.Args[1])
			return nil
		},
	}
}

func gateserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "4",
		Usage: "run 4 node (WebSocket gateway)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			path, _ := getParameters(c)
			gateserver.Run(path, os.Args[1])
			return nil
		},
	}
}

func slgserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "5",
		Usage: "run 5 node (slgserver logic)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			path, _ := getParameters(c)
			slgserver.Run(path, os.Args[1])
			return nil
		},
	}
}

func chatserverCommand() *cli.Command {
	return &cli.Command{
		Name:  "6",
		Usage: "run 6 node (chat server)",
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			path, _ := getParameters(c)
			chatserver.Run(path, os.Args[1])
			return nil
		},
	}
}

func getParameters(c *cli.Context) (path, node string) {
	path = c.String("path")
	node = c.String("node")
	return path, node
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "path",
			Usage:    "profile config file path",
			Required: false,
			Value:    "../config/server/game-cluster.json",
		},
		&cli.StringFlag{
			Name:     "node",
			Usage:    "node id",
			Required: false,
			Value:    "",
		},
	}
}
