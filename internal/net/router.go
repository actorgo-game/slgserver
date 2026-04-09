package net

import (
	"strings"

	clog "github.com/actorgo-game/actorgo/logger"
)

type HandlerFunc func(req *WsMsgReq, rsp *WsMsgRsp)
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type Group struct {
	prefix     string
	hMap       map[string]HandlerFunc
	hMapMidd   map[string][]MiddlewareFunc
	middleware []MiddlewareFunc
}

func (g *Group) AddRouter(name string, handlerFunc HandlerFunc, middleware ...MiddlewareFunc) {
	g.hMap[name] = handlerFunc
	g.hMapMidd[name] = middleware
}

func (g *Group) Use(middleware ...MiddlewareFunc) *Group {
	g.middleware = append(g.middleware, middleware...)
	return g
}

func (g *Group) applyMiddleware(name string) HandlerFunc {
	h, ok := g.hMap[name]
	if !ok {
		h, ok = g.hMap["*"]
	}
	if ok {
		for i := len(g.middleware) - 1; i >= 0; i-- {
			h = g.middleware[i](h)
		}
		if mid, ok2 := g.hMapMidd[name]; ok2 {
			for i := len(mid) - 1; i >= 0; i-- {
				h = mid[i](h)
			}
		}
	}
	return h
}

func (g *Group) exec(name string, req *WsMsgReq, rsp *WsMsgRsp) {
	h := g.applyMiddleware(name)
	if h == nil {
		clog.Warn("Group has not handler: msgName=%s", req.Body.Name)
	} else {
		h(req, rsp)
	}
}

type Router struct {
	groups []*Group
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Group(prefix string) *Group {
	g := &Group{
		prefix:   prefix,
		hMap:     make(map[string]HandlerFunc),
		hMapMidd: make(map[string][]MiddlewareFunc),
	}
	r.groups = append(r.groups, g)
	return g
}

func (r *Router) Run(req *WsMsgReq, rsp *WsMsgRsp) {
	name := req.Body.Name
	msgName := name
	sArr := strings.Split(name, ".")
	prefix := ""
	if len(sArr) == 2 {
		prefix = sArr[0]
		msgName = sArr[1]
	}

	for _, g := range r.groups {
		if g.prefix == prefix {
			g.exec(msgName, req, rsp)
		} else if g.prefix == "*" {
			g.exec(msgName, req, rsp)
		}
	}
}
