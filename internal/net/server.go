package net

import (
	"net/http"

	clog "github.com/actorgo-game/actorgo/logger"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	addr        string
	router      *Router
	needSecret  bool
	beforeClose func(WSConn)
}

func NewServer(addr string, needSecret bool) *Server {
	s := &Server{
		addr:       addr,
		needSecret: needSecret,
	}
	return s
}

func (s *Server) Router(router *Router) {
	s.router = router
}

func (s *Server) Start() {
	clog.Info("ws server starting on %s", s.addr)
	http.HandleFunc("/", s.wsHandler)
	http.ListenAndServe(s.addr, nil)
}

func (s *Server) SetOnBeforeClose(hookFunc func(WSConn)) {
	s.beforeClose = hookFunc
}

func (s *Server) wsHandler(resp http.ResponseWriter, req *http.Request) {
	wsSocket, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}

	conn := ConnMgr.NewConn(wsSocket, s.needSecret)
	clog.Info("client connect: addr=%s", wsSocket.RemoteAddr().String())

	conn.SetRouter(s.router)
	conn.SetOnClose(ConnMgr.RemoveConn)
	conn.SetOnBeforeClose(s.beforeClose)
	conn.Start()
	conn.Handshake()
}
