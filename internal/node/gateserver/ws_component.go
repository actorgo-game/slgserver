package gateserver

import (
	cfacade "github.com/actorgo-game/actorgo/facade"
	clog "github.com/actorgo-game/actorgo/logger"
	mynet "github.com/llr104/slgserver/internal/net"
)

type wsComponent struct {
	cfacade.Component
	server *mynet.Server
}

func NewWSComponent() *wsComponent {
	return &wsComponent{}
}

func (*wsComponent) Name() string {
	return "ws_gate"
}

func (w *wsComponent) OnAfterInit() {
	addr := w.App().Settings().GetString("ws_addr", ":8004")
	needSecret := w.App().Settings().GetBool("need_secret", false)

	w.server = mynet.NewServer(addr, needSecret)
	router := mynet.NewRouter()
	buildRouter(router, w.App())
	w.server.Router(router)
	w.server.SetOnBeforeClose(makeOnConnClose(w.App()))
	go w.server.Start()

	clog.Info("[ws_gate] component started, addr=%s, secret=%v", addr, needSecret)
}

func (w *wsComponent) OnStop() {
	clog.Info("[ws_gate] component stopped")
}
