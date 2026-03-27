package net

import (
	"fmt"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
)

func ElapsedTime() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *WsMsgReq, rsp *WsMsgRsp) {
			bt := time.Now().UnixNano()
			next(req, rsp)
			et := time.Now().UnixNano()
			diff := (et - bt) / int64(time.Millisecond)
			clog.Info("ElapsedTime: msgName=%s, cost=%dms", req.Body.Name, diff)
		}
	}
}

func Log() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *WsMsgReq, rsp *WsMsgRsp) {
			clog.Info("client req: msgName=%s, data=%v", req.Body.Name, fmt.Sprintf("%v", req.Body.Msg))
			next(req, rsp)
		}
	}
}

func CheckLogin() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *WsMsgReq, rsp *WsMsgRsp) {
			_, err := req.Conn.GetProperty("uid")
			if err != nil {
				clog.Warn("connect not found uid: msgName=%s", req.Body.Name)
				rsp.Body.Code = 106
				return
			}
			next(req, rsp)
		}
	}
}

func CheckRole() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *WsMsgReq, rsp *WsMsgRsp) {
			_, err := req.Conn.GetProperty("role")
			if err != nil {
				clog.Warn("connect not found role: msgName=%s", req.Body.Name)
				rsp.Body.Code = 107
				return
			}
			next(req, rsp)
		}
	}
}

func CheckRId() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *WsMsgReq, rsp *WsMsgRsp) {
			_, err := req.Conn.GetProperty("rid")
			if err != nil {
				clog.Warn("connect not found rid: msgName=%s", req.Body.Name)
				rsp.Body.Code = 107
				return
			}
			next(req, rsp)
		}
	}
}
