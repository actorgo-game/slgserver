package gateserver

import (
	"time"

	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	"github.com/llr104/slgserver/internal/rpc"
)

type checkCenterComponent struct {
	cfacade.Component
}

func newCheckCenter() *checkCenterComponent {
	return &checkCenterComponent{}
}

func (*checkCenterComponent) Name() string {
	return "check_center"
}

func (c *checkCenterComponent) Init() {
	for i := 0; i < 30; i++ {
		if rpc.PingCenter(c.App()) {
			clog.Info("[check_center] loginserver node is ready")
			return
		}
		clog.Info("[check_center] waiting for loginserver node... (%d/30)", i+1)
		time.Sleep(2 * time.Second)
	}
	clog.Warn("[check_center] loginserver node not available after 60s, proceeding anyway")
}
