package master

import (
	"fmt"

	"github.com/actorgo-game/actorgo"
)

func Run(profileFilePath, nodeID string) {
	fmt.Printf("profileFilePath: %s, nodeID: %s\n", profileFilePath, nodeID)
	app := actorgo.Configure(profileFilePath, nodeID, false, actorgo.Cluster)
	app.Startup()
}
